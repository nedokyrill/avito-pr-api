package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const testItem = "apple"

func TestContains(t *testing.T) {
	t.Run("item exists in slice", func(t *testing.T) {
		slice := []string{testItem, "banana", "cherry"}
		item := "banana"

		result := Contains(slice, item)

		assert.True(t, result)
	})

	t.Run("item does not exist in slice", func(t *testing.T) {
		slice := []string{testItem, "banana", "cherry"}
		item := "orange"

		result := Contains(slice, item)

		assert.False(t, result)
	})

	t.Run("empty slice", func(t *testing.T) {
		slice := []string{}
		item := testItem

		result := Contains(slice, item)

		assert.False(t, result)
	})

	t.Run("nil slice", func(t *testing.T) {
		var slice []string
		item := testItem

		result := Contains(slice, item)

		assert.False(t, result)
	})

	t.Run("empty string in slice", func(t *testing.T) {
		slice := []string{"", testItem, "banana"}
		item := ""

		result := Contains(slice, item)

		assert.True(t, result)
	})

	t.Run("single element slice - item exists", func(t *testing.T) {
		slice := []string{testItem}
		item := testItem

		result := Contains(slice, item)

		assert.True(t, result)
	})

	t.Run("single element slice - item does not exist", func(t *testing.T) {
		slice := []string{testItem}
		item := "banana"

		result := Contains(slice, item)

		assert.False(t, result)
	})

	t.Run("case sensitivity", func(t *testing.T) {
		slice := []string{"Apple", "Banana", "Cherry"}
		item := testItem

		result := Contains(slice, item)

		assert.False(t, result, "Contains should be case-sensitive")
	})

	t.Run("duplicate items in slice", func(t *testing.T) {
		slice := []string{testItem, "banana", testItem, "cherry", testItem}
		item := testItem

		result := Contains(slice, item)

		assert.True(t, result)
	})

	t.Run("slice with spaces", func(t *testing.T) {
		slice := []string{testItem + " ", " banana", " cherry "}
		item := testItem

		result := Contains(slice, item)

		assert.False(t, result, "Should not trim spaces")
	})
}
