/*
* Copyright 2021 by EYWA chain <blockchain@digiu.ai>
*/

package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenExportBlocksFileName(t *testing.T) {
	name := "blocks.dat"
	start := uint32(0)
	end := uint32(100)
	fileName := GenExportBlocksFileName(name, start, end)
	assert.Equal(t, "blocks_0_100.dat", fileName)
	name = "blocks"
	fileName = GenExportBlocksFileName(name, start, end)
	assert.Equal(t, "blocks_0_100", fileName)
	name = "blocks."
	fileName = GenExportBlocksFileName(name, start, end)
	assert.Equal(t, "blocks_0_100.", fileName)
	name = "blocks.export.dat"
	fileName = GenExportBlocksFileName(name, start, end)
	assert.Equal(t, "blocks.export_0_100.dat", fileName)
}
