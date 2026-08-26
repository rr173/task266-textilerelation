package model

import (
	"fmt"
	"strings"
)

// ValidateBatch 校验批次输入。
func ValidateBatch(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: batch name is required", ErrInvalid)
	}
	return nil
}

// ValidateSample 校验样本结构输入，重点拦截纹样坐标越界。
func ValidateSample(batchID int64, name, provenance string, warpCount, weftCount, warpDensity, weftDensity int) error {
	if batchID <= 0 {
		return fmt.Errorf("%w: batch id is required", ErrInvalid)
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: sample name is required", ErrInvalid)
	}
	if warpCount < 4 || weftCount < 4 {
		return fmt.Errorf("%w: warp/weft count must be >= 4 (got %d x %d)", ErrInvalid, warpCount, weftCount)
	}
	if warpDensity <= 0 || weftDensity <= 0 {
		return fmt.Errorf("%w: warp/weft density must be positive", ErrInvalid)
	}
	return nil
}

// ValidateMotif 校验纹样单元：坐标与尺寸必须落在样本经纬范围内。
func ValidateMotif(sample *FabricSample, name string, originX, originY, width, height int, grid string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: motif name is required", ErrInvalid)
	}
	if originX < 0 || originY < 0 {
		return fmt.Errorf("%w: motif origin must be non-negative", ErrInvalid)
	}
	if width < 2 || height < 2 {
		return fmt.Errorf("%w: motif width/height must be >= 2", ErrInvalid)
	}
	if originX+width > sample.WarpCount || originY+height > sample.WeftCount {
		return fmt.Errorf("%w: motif box (%d,%d)+%dx%d out of sample %dx%d",
			ErrInvalid, originX, originY, width, height, sample.WarpCount, sample.WeftCount)
	}
	if err := ValidateGridString(grid, width, height); err != nil {
		return err
	}
	return nil
}

// ValidateGridString 校验网格编码：必须恰好 width 列 x height 行，且只含 '#' 与 '.'。
func ValidateGridString(grid string, width, height int) error {
	rows := strings.Split(grid, ",")
	if len(rows) != height {
		return fmt.Errorf("%w: grid has %d rows, want %d", ErrInvalid, len(rows), height)
	}
	for i, row := range rows {
		if len(row) != width {
			return fmt.Errorf("%w: grid row %d has %d cols, want %d", ErrInvalid, i, len(row), width)
		}
		for _, ch := range row {
			if ch != '#' && ch != '.' {
				return fmt.Errorf("%w: grid contains invalid char %q", ErrInvalid, ch)
			}
		}
	}
	return nil
}

// ValidateTechnique 校验工艺特征：织法、浮长、色牢度等。
func ValidateTechnique(weaveClass, interlacing, dyeClass string, colorfastness int) error {
	switch weaveClass {
	case "plain", "twill", "satin", "compound":
	default:
		return fmt.Errorf("%w: unknown weave class %q", ErrInvalid, weaveClass)
	}
	if interlacing == "" {
		return fmt.Errorf("%w: interlacing rule is required", ErrInvalid)
	}
	switch dyeClass {
	case "natural", "synthetic":
	default:
		return fmt.Errorf("%w: unknown dye class %q", ErrInvalid, dyeClass)
	}
	if colorfastness < 1 || colorfastness > 5 {
		return fmt.Errorf("%w: colorfastness must be 1..5", ErrInvalid)
	}
	return nil
}
