// 民族织物纹样工艺关系复核台服务入口。
//
// 用法：
//   task266-textilerelation --addr :8080 --db textile.db
//   task266-textilerelation --smoke-test --db smoke.db
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"task266-textilerelation/internal/httpapi"
	"task266-textilerelation/internal/model"
	"task266-textilerelation/internal/sample"
	"task266-textilerelation/internal/service"
	"task266-textilerelation/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	dbPath := flag.String("db", "textile.db", "SQLite database path")
	smoke := flag.Bool("smoke-test", false, "run end-to-end self test and exit")
	flag.Parse()

	if *smoke {
		if err := runSmokeTest(*dbPath); err != nil {
			log.Fatalf("smoke test failed: %v", err)
		}
		fmt.Println("smoke test passed")
		return
	}

	st, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer st.Close()

	svc := service.New(st)
	srv := &http.Server{
		Addr:              *addr,
		Handler:           httpapi.New(svc).Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("task266-textilerelation listening on %s (db=%s)", *addr, *dbPath)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

// runSmokeTest 端到端自检（--smoke-test 契约的唯一判据）：
//
//  1. 建批次、导入三个织物样本（A 平纹 1/1 / B 斜纹 3/1 / C 斜纹 3/1）与纹样单元；
//  2. 解析单元拓扑（周期+对称），写入工艺特征；
//  3. 比较 A vs B：外观相似但经纬交错规则相反 → 系统判 conflict（视觉巧合）；
//  4. 复核冲突确认否决；比较 B vs C（同工艺同染料）→ confirmed；
//  5. 出处循环检测：C vs B 再确认后建版本 → 拒绝；修正后建版本并 share/freeze；
//  6. 错误边界：坐标越界、重复导入幂等、冻结后再次冻结被拒；
//  7. 关闭数据库并重开同一文件，验证批次/样本/单元/关系/版本全部持久化恢复。
func runSmokeTest(dbPath string) error {
	_ = os.Remove(dbPath)
	st, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	svc := service.New(st)

	// 1) 批次 + 样本导入（哈希幂等）
	batch, err := svc.CreateBatch("merovingian-weaves", "梅罗文加时期织造比对批次")
	if err != nil {
		return fmt.Errorf("create batch: %w", err)
	}

	gridA := "##.#,.##.,##.#,.##."
	gridB := "##.#,.###,##.#,.###"
	gridC := "##.#,.###,##.#,.###"

	sampleA, created, err := svc.ImportSample(sample.IngestInput{
		BatchID: batch.ID, Name: "fragment-a", Provenance: "圣但尼修道院",
		WarpCount: 12, WeftCount: 12, WarpDensity: 18, WeftDensity: 16,
		MotifGrids: map[string]string{"diamond-a": gridA},
	})
	if err != nil {
		return fmt.Errorf("import a: %w", err)
	}
	if !created {
		return fmt.Errorf("expected sample A created")
	}
	// 幂等：重复导入应返回既有样本
	again, created, err := svc.ImportSample(sample.IngestInput{
		BatchID: batch.ID, Name: "fragment-a", Provenance: "圣但尼修道院",
		WarpCount: 12, WeftCount: 12, WarpDensity: 18, WeftDensity: 16,
		MotifGrids: map[string]string{"diamond-a": gridA},
	})
	if err != nil {
		return fmt.Errorf("reimport a: %w", err)
	}
	if created || again.ID != sampleA.ID {
		return fmt.Errorf("expected idempotent reimport, got created=%v id=%d", created, again.ID)
	}

	sampleB, _, err := svc.ImportSample(sample.IngestInput{
		BatchID: batch.ID, Name: "fragment-b", Provenance: "图尔主教座堂",
		WarpCount: 12, WeftCount: 12, WarpDensity: 24, WeftDensity: 22,
		MotifGrids: map[string]string{"diamond-b": gridB},
	})
	if err != nil {
		return fmt.Errorf("import b: %w", err)
	}
	sampleC, _, err := svc.ImportSample(sample.IngestInput{
		BatchID: batch.ID, Name: "fragment-c", Provenance: "图尔主教座堂",
		WarpCount: 12, WeftCount: 12, WarpDensity: 24, WeftDensity: 22,
		MotifGrids: map[string]string{"diamond-c": gridC},
	})
	if err != nil {
		return fmt.Errorf("import c: %w", err)
	}

	// 2) 工艺特征：A 平纹 1/1 茜草；B/C 斜纹 3/1 靛蓝
	techs := []*model.TechniqueFeature{
		{SampleID: sampleA.ID, WeaveClass: model.WeavePlain, Interlacing: model.RulePlain11,
			DyeClass: "natural", DyePigment: "茜草", Colorfastness: 4, CarbonRatio: -18.5},
		{SampleID: sampleB.ID, WeaveClass: model.WeaveTwill, Interlacing: model.RuleTwill31,
			TwillDirection: "right", DyeClass: "natural", DyePigment: "靛蓝", Colorfastness: 4, CarbonRatio: -25.0},
		{SampleID: sampleC.ID, WeaveClass: model.WeaveTwill, Interlacing: model.RuleTwill31,
			TwillDirection: "right", DyeClass: "natural", DyePigment: "靛蓝", Colorfastness: 5, CarbonRatio: -24.8},
	}
	for _, t := range techs {
		if _, err := svc.UpsertTechnique(t); err != nil {
			return fmt.Errorf("technique sample %d: %w", t.SampleID, err)
		}
	}

	// 3) 解析单元拓扑
	motifsA, err := svc.ListMotifs(sampleA.ID)
	if err != nil || len(motifsA) != 1 {
		return fmt.Errorf("motifs a: %v len=%d", err, len(motifsA))
	}
	motifsB, err := svc.ListMotifs(sampleB.ID)
	if err != nil || len(motifsB) != 1 {
		return fmt.Errorf("motifs b: %v len=%d", err, len(motifsB))
	}
	motifsC, err := svc.ListMotifs(sampleC.ID)
	if err != nil || len(motifsC) != 1 {
		return fmt.Errorf("motifs c: %v len=%d", err, len(motifsC))
	}
	unitA, err := svc.ParseMotif(motifsA[0].ID)
	if err != nil {
		return fmt.Errorf("parse a: %w", err)
	}
	unitB, err := svc.ParseMotif(motifsB[0].ID)
	if err != nil {
		return fmt.Errorf("parse b: %w", err)
	}
	unitC, err := svc.ParseMotif(motifsC[0].ID)
	if err != nil {
		return fmt.Errorf("parse c: %w", err)
	}
	if unitA.PeriodY == 0 || unitB.PeriodY == 0 {
		return fmt.Errorf("expected periods parsed, a=(%d,%d) b=(%d,%d)",
			unitA.PeriodX, unitA.PeriodY, unitB.PeriodX, unitB.PeriodY)
	}

	// 4) A vs B：外观相似但交错规则相反 → conflict
	relAB, err := svc.CompareMotifs(unitA.ID, unitB.ID)
	if err != nil {
		return fmt.Errorf("compare a-b: %w", err)
	}
	if relAB.Verdict != model.VerdictConflict {
		return fmt.Errorf("expected A-B conflict (visual coincidence), got %s (sim=%.2f weave=%s)",
			relAB.Verdict, relAB.TopoSimilarity, relAB.WeaveCompat)
	}
	// 复核确认否决
	relAB, err = svc.ResolveConflictVerdict(relAB.ID, true)
	if err != nil {
		return fmt.Errorf("resolve conflict: %w", err)
	}
	if relAB.Verdict != model.VerdictRejected {
		return fmt.Errorf("expected A-B rejected after review, got %s", relAB.Verdict)
	}

	// 5) B vs C：同工艺同染料 → confirmed
	relBC, err := svc.CompareMotifs(unitB.ID, unitC.ID)
	if err != nil {
		return fmt.Errorf("compare b-c: %w", err)
	}
	if relBC.Verdict != model.VerdictConfirmed {
		return fmt.Errorf("expected B-C confirmed, got %s (sim=%.2f weave=%s dye=%s)",
			relBC.Verdict, relBC.TopoSimilarity, relBC.WeaveCompat, relBC.DyeCompat)
	}

	// 6) 出处循环检测：C vs B 也确认 → 关系图成环 → 建版本必须拒绝
	relCB, err := svc.CompareMotifs(unitC.ID, unitB.ID)
	if err != nil {
		return fmt.Errorf("compare c-b: %w", err)
	}
	if relCB.Verdict != model.VerdictConfirmed {
		return fmt.Errorf("expected C-B confirmed, got %s", relCB.Verdict)
	}
	if _, err := svc.CreateVersion("v1", "含环版本"); err == nil {
		return fmt.Errorf("expected provenance cycle rejection")
	}
	// 修正：否决 C-B，再建版本 → 成功
	if _, err := svc.ApplyVerdict(relCB.ID, model.VerdictRejected, "环修正"); err != nil {
		return fmt.Errorf("reject c-b: %w", err)
	}
	version, err := svc.CreateVersion("v1", "B-C 工艺传承确认")
	if err != nil {
		return fmt.Errorf("create version: %w", err)
	}

	// 7) share → freeze（含反证快照）
	if _, err := svc.ShareVersion(version.ID); err != nil {
		return fmt.Errorf("share version: %w", err)
	}
	frozen, err := svc.FreezeVersion(version.ID)
	if err != nil {
		return fmt.Errorf("freeze version: %w", err)
	}
	if frozen.Status != model.VersionFrozen || frozen.ProvenanceNote == "" {
		return fmt.Errorf("expected frozen with provenance note, got %s note=%q",
			frozen.Status, frozen.ProvenanceNote)
	}
	// 冻结后再次冻结 → 拒绝
	if _, err := svc.FreezeVersion(version.ID); err == nil {
		return fmt.Errorf("expected refreeze rejection")
	}

	// 8) 错误边界：坐标越界
	if _, err := svc.AddMotif(sampleA.ID, "out-of-range", 100, 100, 4, 4, gridA); err == nil {
		return fmt.Errorf("expected out-of-range rejection")
	}
	// 反证：对已收录关系追加补片反证（不自动改写已入版本的关系，但保留记录）
	if _, err := svc.AddCounterEvidence(relBC.ID, "patch", "B 单元疑为后期补片", "档案-88"); err != nil {
		return fmt.Errorf("add evidence: %w", err)
	}
	evs, err := svc.ListEvidence(relBC.ID)
	if err != nil || len(evs) != 1 {
		return fmt.Errorf("evidence count: %v len=%d", err, len(evs))
	}

	// 9) 批次状态机推进到 published
	for i := 0; i < 3; i++ {
		if _, err := svc.AdvanceBatch(batch.ID); err != nil {
			return fmt.Errorf("advance batch step %d: %w", i, err)
		}
	}
	publishedBatch, err := svc.GetBatch(batch.ID)
	if err != nil {
		return err
	}
	if publishedBatch.Status != model.BatchPublished {
		return fmt.Errorf("expected published batch, got %s", publishedBatch.Status)
	}

	// 10) 关闭并重开：验证重启恢复
	savedIDs := []int64{batch.ID, sampleA.ID, sampleB.ID, sampleC.ID, unitA.ID, relAB.ID, relBC.ID, version.ID}
	savedBatchStatus, savedRelVerdict, savedVersionStatus := publishedBatch.Status, relBC.Verdict, frozen.Status
	if err := st.Close(); err != nil {
		return err
	}
	st2, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("reopen: %w", err)
	}
	defer st2.Close()
	svc2 := service.New(st2)

	if b, err := svc2.GetBatch(savedIDs[0]); err != nil || b.Status != savedBatchStatus {
		return fmt.Errorf("batch recovery: %v status=%s", err, b.Status)
	}
	if _, err := svc2.GetSample(savedIDs[1]); err != nil {
		return fmt.Errorf("sample recovery: %w", err)
	}
	if _, err := svc2.GetMotif(savedIDs[4]); err != nil {
		return fmt.Errorf("motif recovery: %w", err)
	}
	relBC2, _, err := svc2.GetRelation(savedIDs[6])
	if err != nil {
		return fmt.Errorf("relation recovery: %w", err)
	}
	if relBC2.Verdict != savedRelVerdict {
		return fmt.Errorf("relation verdict not recovered: %s", relBC2.Verdict)
	}
	v2, _, err := svc2.GetVersion(savedIDs[7])
	if err != nil {
		return fmt.Errorf("version recovery: %w", err)
	}
	if v2.Status != savedVersionStatus || len(v2.RelationIDs) == 0 {
		return fmt.Errorf("version not recovered: status=%s ids=%q", v2.Status, v2.RelationIDs)
	}
	rels, err := svc2.ListRelations("")
	if err != nil || len(rels) != 3 {
		return fmt.Errorf("relations recovery: %v len=%d", err, len(rels))
	}
	stats, err := svc2.GetStats()
	if err != nil {
		return err
	}
	fmt.Printf("smoke: batch=%d samples=%d motifs=%d relations=%d versions=%d batch_status=%s relBC=%s\n",
		stats.Batches, stats.Samples, stats.Motifs, len(rels), stats.Versions,
		savedBatchStatus, savedRelVerdict)
	return nil
}
