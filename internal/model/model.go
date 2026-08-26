// Package model 定义民族织物纹样工艺关系复核台的核心实体。
//
// 实体覆盖纺织史研究证据链的四层对象：
//   - 样本批次（Batch）：导入工作的组织单元，带自身状态机；
//   - 织物样本（FabricSample）：一块实物的数字化摘要，含经纬结构与内容哈希；
//   - 纹样单元（MotifUnit）：样本上被解析出的重复装饰单元，携带网格拓扑；
//   - 工艺特征（TechniqueFeature）：织造技法与染料证据，用于验证传承关系。
package model

import "time"

// Batch 是一次织物样本导入/复核工作的组织单元。
type Batch struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Note      string    `json:"note"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// FabricSample 是单件织物样本的数字化摘要。
//
// SHA256 为幂等键：相同结构内容重复导入时返回既有样本。
type FabricSample struct {
	ID          int64     `json:"id"`
	BatchID     int64     `json:"batch_id"`
	Name        string    `json:"name"`
	Provenance  string    `json:"provenance"` // 出土/馆藏/传承出处
	WarpCount   int       `json:"warp_count"` // 经线根数（全幅或取样区）
	WeftCount   int       `json:"weft_count"` // 纬线根数
	WarpDensity int       `json:"warp_density"` // 经密：根/cm
	WeftDensity int       `json:"weft_density"` // 纬密：根/cm
	SHA256      string    `json:"sha256"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// MotifUnit 是织物上的一个纹样单元。
//
// GridPattern 用 '#' 表示浮色格、'.' 表示底色格，行间以 ',' 分隔；
// PeriodX/PeriodY 与 Symmetry 由拓扑解析器计算并回填。
type MotifUnit struct {
	ID        int64     `json:"id"`
	SampleID  int64     `json:"sample_id"`
	Name      string    `json:"name"`
	OriginX   int       `json:"origin_x"` // 单元左上角经向坐标
	OriginY   int       `json:"origin_y"` // 单元左上角纬向坐标
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	Grid      string    `json:"grid"`
	PeriodX   int       `json:"period_x"`
	PeriodY   int       `json:"period_y"`
	Symmetry  string    `json:"symmetry"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// TechniqueFeature 记录样本的织造技法与染料证据。
type TechniqueFeature struct {
	ID             int64   `json:"id"`
	SampleID       int64   `json:"sample_id"`
	WeaveClass     string  `json:"weave_class"`     // plain/twill/satin/compound
	Interlacing    string  `json:"interlacing"`     // 浮长规则，如 1/1、2/1、5/2
	TwillDirection string  `json:"twill_direction"` // left/right
	YarnTwist      string  `json:"yarn_twist"`      // S/Z 捻向
	DyeClass       string  `json:"dye_class"`       // natural/synthetic
	DyePigment     string  `json:"dye_pigment"`     // 茜草/靛蓝/黄栌/苏木/...
	Colorfastness  int     `json:"colorfastness"`   // 色牢度 1~5 级
	CarbonRatio    float64 `json:"carbon_ratio"`    // 碳同位素比值 δ13C
	UpdatedAt      time.Time `json:"updated_at"`
}
