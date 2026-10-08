package domain_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/netologist/hotel-booking/internal/domain"
)

const (
	testValidName       = "Grand Budapest Hotel"
	testValidAddress    = "100 Alpine Way"
	testValidCity       = "Lutz"
	errExpectedNameLen  = "name length must be between 2 and 255"
	errExpectedAddrLen  = "address length must be between 2 and 255"
	errExpectedCityLen  = "city length must be between 2 and 100"
	testWhitespaceValue = "   "
)

type invariantTestCase struct {
	name        string
	hotelName   string
	address     string
	city        string
	errContains string
}

func nameViolationCases() []invariantTestCase {
	return []invariantTestCase{
		{
			name:        "empty name",
			hotelName:   "",
			address:     testValidAddress,
			city:        testValidCity,
			errContains: errExpectedNameLen,
		},
		{
			name:        "whitespace-only name",
			hotelName:   testWhitespaceValue,
			address:     testValidAddress,
			city:        testValidCity,
			errContains: errExpectedNameLen,
		},
		{
			name:        "single char name",
			hotelName:   "A",
			address:     testValidAddress,
			city:        testValidCity,
			errContains: errExpectedNameLen,
		},
		{
			name:        "name exceeds max length",
			hotelName:   strings.Repeat("a", 256),
			address:     testValidAddress,
			city:        testValidCity,
			errContains: errExpectedNameLen,
		},
	}
}

func addressViolationCases() []invariantTestCase {
	return []invariantTestCase{
		{
			name:        "empty address",
			hotelName:   testValidName,
			address:     "",
			city:        testValidCity,
			errContains: errExpectedAddrLen,
		},
		{
			name:        "whitespace-only address",
			hotelName:   testValidName,
			address:     testWhitespaceValue,
			city:        testValidCity,
			errContains: errExpectedAddrLen,
		},
		{
			name:        "single char address",
			hotelName:   testValidName,
			address:     "X",
			city:        testValidCity,
			errContains: errExpectedAddrLen,
		},
		{
			name:        "address exceeds max length",
			hotelName:   testValidName,
			address:     strings.Repeat("b", 256),
			city:        testValidCity,
			errContains: errExpectedAddrLen,
		},
	}
}

func cityViolationCases() []invariantTestCase {
	return []invariantTestCase{
		{
			name:        "empty city",
			hotelName:   testValidName,
			address:     testValidAddress,
			city:        "",
			errContains: errExpectedCityLen,
		},
		{
			name:        "whitespace-only city",
			hotelName:   testValidName,
			address:     testValidAddress,
			city:        testWhitespaceValue,
			errContains: errExpectedCityLen,
		},
		{
			name:        "single char city",
			hotelName:   testValidName,
			address:     testValidAddress,
			city:        "L",
			errContains: errExpectedCityLen,
		},
		{
			name:        "city exceeds max length",
			hotelName:   testValidName,
			address:     testValidAddress,
			city:        strings.Repeat("c", 101),
			errContains: errExpectedCityLen,
		},
	}
}

func TestNewHotel_Success(t *testing.T) {
	hotel, err := domain.NewHotel("  Grand Budapest Hotel  ", " 100 Alpine Way ", " Lutz ")
	require.NoError(t, err)

	assert.NotEqual(t, uuid.Nil, hotel.ID)
	assert.Equal(t, testValidName, hotel.Name)
	assert.Equal(t, testValidAddress, hotel.Address)
	assert.Equal(t, testValidCity, hotel.City)
}

func TestNewHotel_TrimmingInvariant(t *testing.T) {
	hotel, err := domain.NewHotel("\tHotel California\n", "  Sunset Blvd  ", "  Los Angeles  ")
	require.NoError(t, err)

	assert.Equal(t, "Hotel California", hotel.Name)
	assert.Equal(t, "Sunset Blvd", hotel.Address)
	assert.Equal(t, "Los Angeles", hotel.City)
}

func TestNewHotel_InvariantViolations(t *testing.T) {
	testCases := append(nameViolationCases(), addressViolationCases()...)
	testCases = append(testCases, cityViolationCases()...)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hotel, err := domain.NewHotel(tc.hotelName, tc.address, tc.city)
			require.Error(t, err)
			require.ErrorIs(t, err, domain.ErrInvalidHotel)
			assert.Contains(t, err.Error(), tc.errContains)
			assert.Zero(t, hotel)
		})
	}
}

func TestNewHotel_UnicodeBoundary(t *testing.T) {
	hotel, err := domain.NewHotel("İstanbul Otel", "İstiklal Cad. No:1", "İstanbul")
	require.NoError(t, err)
	assert.Equal(t, "İstanbul Otel", hotel.Name)
	assert.Equal(t, "İstiklal Cad. No:1", hotel.Address)
	assert.Equal(t, "İstanbul", hotel.City)
}
