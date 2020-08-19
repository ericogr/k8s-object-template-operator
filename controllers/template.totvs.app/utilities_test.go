package controllers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddIdentationBasic(t *testing.T) {
	textToIdent := `
spec:
  image: image-name
  tag: \"v1.2\"
  params:
    param1: value1
    param2: value2`

	textIdented := addIdentation(textToIdent)
	textExpected := `
  spec:
    image: image-name
    tag: \"v1.2\"
    params:
      param1: value1
      param2: value2`

	require.Equal(t, textExpected, textIdented)
}
