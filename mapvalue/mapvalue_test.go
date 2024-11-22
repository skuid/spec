package mapvalue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsString(t *testing.T) {
	places := map[string]interface{}{
		"taco mamacita": "north shore",
		"stir":          "southside",
		"chili's":       []string{"downtown", "near the mall"},
	}
	cases := []struct {
		testDescription string
		source          map[string]interface{}
		key             string
		wantErrorMsg    string
	}{
		{
			"Should return nil error when value found as string",
			places,
			"stir",
			"",
		},
		{
			"Should return error about missing key when value missing",
			places,
			"community pie",
			"community pie must be provided",
		},
		{
			"Should return error about bad type when value exists with wrong type",
			places,
			"chili's",
			"chili's found with wrong type: expected string",
		},
	}
	for _, c := range cases {
		err := IsString(c.source, c.key)
		if c.wantErrorMsg == "" {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, c.wantErrorMsg)
		}
	}
}

func TestString(t *testing.T) {
	places := map[string]interface{}{
		"taco mamacita": "north shore",
		"stir":          "southside",
		"chili's":       []string{"downtown", "near the mall"},
	}
	cases := []struct {
		testDescription string
		source          map[string]interface{}
		key             string
		wantReturn      string
	}{
		{
			"Should return value when value found as string",
			places,
			"stir",
			"southside",
		},
		{
			"Should return empty string when value missing",
			places,
			"community pie",
			"",
		},
		{
			"Should return empty string when value exists with wrong type",
			places,
			"chili's",
			"",
		},
	}
	for _, c := range cases {
		returned := String(c.source, c.key)
		assert.Equal(t, returned, c.wantReturn)
	}
}

func TestIsMapSlice(t *testing.T) {
	places := map[string]interface{}{
		"restaurants": "main street meats",
		"bars":        []interface{}{map[string]interface{}{"stir": "southside"}, map[string]interface{}{"mike's": "northshore"}},
		"gyms":        []interface{}{"non map interface", map[string]interface{}{"sportsbarn": "downtown"}},
	}
	cases := []struct {
		testDescription string
		source          map[string]interface{}
		key             string
		wantErrorMsg    string
	}{
		{
			"Should return nil error when value found as map slice",
			places,
			"bars",
			"",
		},
		{
			"Should return error about missing key when value missing",
			places,
			"offices",
			"offices must be provided",
		},
		{
			"Should return error about bad type when value exists with wrong type",
			places,
			"restaurants",
			"restaurants found with wrong type: expected JSON array",
		},
		{
			"Should return error about bad type when nested value exists with wrong type",
			places,
			"gyms",
			"Object at index 0 in array found with wrong type: expected JSON object",
		},
	}
	for _, c := range cases {
		err := IsMapSlice(c.source, c.key)
		if c.wantErrorMsg == "" {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, c.wantErrorMsg)
		}
	}
}

func TestMapSlice(t *testing.T) {
	places := map[string]interface{}{
		"restaurants": "main street meats",
		"bars":        []interface{}{map[string]interface{}{"stir": "southside"}, map[string]interface{}{"mike's": "northshore"}},
		"gyms":        []interface{}{"non map interface", map[string]interface{}{"sportsbarn": "downtown"}},
	}
	cases := []struct {
		testDescription string
		source          map[string]interface{}
		key             string
		wantReturn      []map[string]interface{}
	}{
		{
			"Should return value when value found as map slice",
			places,
			"bars",
			[]map[string]interface{}{{"stir": "southside"}, {"mike's": "northshore"}},
		},
		{
			"Should return empty slice when value missing",
			places,
			"offices",
			[]map[string]interface{}{},
		},
		{
			"Should return empty slice when value exists with wrong type",
			places,
			"restaurants",
			[]map[string]interface{}{},
		},
		{
			"Should return empty slice when nested value exists with wrong type",
			places,
			"gyms",
			[]map[string]interface{}{},
		},
	}
	for _, c := range cases {
		returned := MapSlice(c.source, c.key)
		assert.Equal(t, returned, c.wantReturn)
	}
}

func TestIsMap(t *testing.T) {
	places := map[string]interface{}{
		"restaurants": "main street meats",
		"bars":        map[string]interface{}{"stir": "southside"},
	}
	cases := []struct {
		testDescription string
		source          map[string]interface{}
		key             string
		wantErrorMsg    string
	}{
		{
			"Should return nil error when value found as map",
			places,
			"bars",
			"",
		},
		{
			"Should return error about missing key when value missing",
			places,
			"offices",
			"offices must be provided",
		},
		{
			"Should return error about bad type when value exists with wrong type",
			places,
			"restaurants",
			"restaurants found with wrong type: expected JSON object",
		},
	}
	for _, c := range cases {
		err := IsMap(c.source, c.key)
		if c.wantErrorMsg == "" {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, c.wantErrorMsg)
		}
	}
}

func TestMap(t *testing.T) {
	places := map[string]interface{}{
		"restaurants": "main street meats",
		"bars":        map[string]interface{}{"stir": "southside"},
	}
	cases := []struct {
		testDescription string
		source          map[string]interface{}
		key             string
		wantReturn      map[string]interface{}
	}{
		{
			"Should return value when value found as map",
			places,
			"bars",
			map[string]interface{}{"stir": "southside"},
		},
		{
			"Should return empty map when value missing",
			places,
			"offices",
			map[string]interface{}{},
		},
		{
			"Should return empty slice when value exists with wrong type",
			places,
			"restaurants",
			map[string]interface{}{},
		},
	}
	for _, c := range cases {
		returned := Map(c.source, c.key)
		assert.Equal(t, returned, c.wantReturn)
	}
}

type CombineMap struct {
	Quantity int
	Price    float64
	Name     string
}

func TestCombineStructWithMap(t *testing.T) {
	t.Run("should combine a struct with a map", func(t *testing.T) {
		inputStruct := CombineMap{
			12,
			19.00,
			"sock",
		}
		inputMap := map[string]interface{}{
			"Quantity": 1,
			"Name":     "box",
		}
		wantStruct := CombineMap{
			1,
			19.00,
			"box",
		}

		outputStruct, err := CombineStructWithMap(inputStruct, inputMap)
		assert.Equal(t, wantStruct, outputStruct)
		assert.Equal(t, nil, err)
	})
}
