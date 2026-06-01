package generator

import (
	"context"
	"strconv"
	"strings"
	"testing"
)

func TestCountryCodeGenerator(t *testing.T) {
	g := &CountryCodeGenerator{}
	ctx := context.Background()
	row := testRow{}

	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		code := v.(string)
		if len(code) != 2 {
			t.Errorf("expected 2-char country code, got %d: %q", len(code), code)
		}
		seen[code] = true
	}
	if len(seen) < 5 {
		t.Errorf("expected at least 5 different country codes, got %d", len(seen))
	}
}

func TestLatitudeGenerator(t *testing.T) {
	g := &LatitudeGenerator{}
	ctx := context.Background()
	row := testRow{}

	for i := 0; i < 100; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		lat := v.(string)
		f, err := strconv.ParseFloat(lat, 64)
		if err != nil {
			t.Fatalf("invalid latitude: %q: %v", lat, err)
		}
		if f < -90 || f > 90 {
			t.Errorf("latitude %f out of range [-90, 90]", f)
		}
	}
}

func TestLongitudeGenerator(t *testing.T) {
	g := &LongitudeGenerator{}
	ctx := context.Background()
	row := testRow{}

	for i := 0; i < 100; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		lon := v.(string)
		f, err := strconv.ParseFloat(lon, 64)
		if err != nil {
			t.Fatalf("invalid longitude: %q: %v", lon, err)
		}
		if f < -180 || f > 180 {
			t.Errorf("longitude %f out of range [-180, 180]", f)
		}
	}
}

func TestPostalCodeGenerator(t *testing.T) {
	g := &PostalCodeGenerator{}
	ctx := context.Background()
	row := testRow{}

	for i := 0; i < 50; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		code := v.(string)
		if len(code) != 5 {
			t.Errorf("expected 5-char postal code, got %d: %q", len(code), code)
		}
	}
}

func TestCityGeneratorReturnsKnownCity(t *testing.T) {
	g := &CityGenerator{}
	ctx := context.Background()
	row := testRow{}

	v, _ := g.Generate(ctx, row)
	city := v.(string)
	if len(city) == 0 {
		t.Error("empty city")
	}
	if strings.Contains(city, " ") && len(city) > 20 {
		t.Errorf("suspiciously long city: %q", city)
	}
}

func TestCountryGeneratorReturnsKnownCountry(t *testing.T) {
	g := &CountryGenerator{}
	ctx := context.Background()
	row := testRow{}

	v, _ := g.Generate(ctx, row)
	country := v.(string)
	if len(country) == 0 {
		t.Error("empty country")
	}
}
