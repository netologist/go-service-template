package domain_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/netologist/go-service-template/internal/domain"
)

const (
	testValidName        = "Mechanical Keyboard"
	testValidDescription = "Ergonomic hot-swappable mechanical keyboard"
	testValidPrice       = int64(9900)
	errExpectedNameLen   = "name length must be between 2 and 255"
	errExpectedDescLen   = "description length must not exceed 1000"
	errExpectedPriceNeg  = "price must be non-negative"
	testWhitespaceValue  = "   "
)

type invariantTestCase struct {
	name        string
	itemName    string
	description string
	price       int64
	errContains string
}

func nameViolationCases() []invariantTestCase {
	return []invariantTestCase{
		{
			name:        "rejects empty name",
			itemName:    "",
			description: testValidDescription,
			price:       testValidPrice,
			errContains: errExpectedNameLen,
		},
		{
			name:        "rejects single character name",
			itemName:    "A",
			description: testValidDescription,
			price:       testValidPrice,
			errContains: errExpectedNameLen,
		},
		{
			name:        "rejects name with only whitespace",
			itemName:    testWhitespaceValue,
			description: testValidDescription,
			price:       testValidPrice,
			errContains: errExpectedNameLen,
		},
		{
			name:        "rejects name exceeding 255 characters",
			itemName:    strings.Repeat("a", domain.MaxNameLength+1),
			description: testValidDescription,
			price:       testValidPrice,
			errContains: errExpectedNameLen,
		},
	}
}

func descAndPriceViolationCases() []invariantTestCase {
	return []invariantTestCase{
		{
			name:        "rejects description exceeding 1000 characters",
			itemName:    testValidName,
			description: strings.Repeat("d", domain.MaxDescriptionLength+1),
			price:       testValidPrice,
			errContains: errExpectedDescLen,
		},
		{
			name:        "rejects negative price",
			itemName:    testValidName,
			description: testValidDescription,
			price:       -1,
			errContains: errExpectedPriceNeg,
		},
	}
}

func TestNewItem_Success(t *testing.T) {
	t.Parallel()

	item, err := domain.NewItem(testValidName, testValidDescription, testValidPrice)

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, item.ID)
	assert.Equal(t, testValidName, item.Name)
	assert.Equal(t, testValidDescription, item.Description)
	assert.Equal(t, testValidPrice, item.Price)
}

func TestNewItem_TrimmingInvariant(t *testing.T) {
	t.Parallel()

	item, err := domain.NewItem("  "+testValidName+"  ", "  "+testValidDescription+"  ", testValidPrice)

	require.NoError(t, err)
	assert.Equal(t, testValidName, item.Name)
	assert.Equal(t, testValidDescription, item.Description)
}

func TestNewItem_InvariantViolations(t *testing.T) {
	t.Parallel()

	allCases := append(nameViolationCases(), descAndPriceViolationCases()...)

	for _, tc := range allCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			item, err := domain.NewItem(tc.itemName, tc.description, tc.price)

			require.Error(t, err)
			require.ErrorIs(t, err, domain.ErrInvalidItem)
			assert.Contains(t, err.Error(), tc.errContains)
			assert.Equal(t, domain.Item{}, item)
		})
	}
}

func TestNewItem_UnicodeBoundary(t *testing.T) {
	t.Parallel()

	item, err := domain.NewItem("Üç", "Özet", 0)

	require.NoError(t, err)
	assert.Equal(t, "Üç", item.Name)
	assert.Equal(t, "Özet", item.Description)
	assert.Equal(t, int64(0), item.Price)
}
