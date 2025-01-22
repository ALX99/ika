package requestid

import (
	"context"
	"crypto/rand"
	rrand "math/rand"
	rrand2 "math/rand/v2"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alx99/ika"
	"github.com/google/uuid"
	"github.com/matryer/is"
	"github.com/segmentio/ksuid"
)

func TestModifyRequest_uuidv7(t *testing.T) {
	is := is.New(t)
	t.Parallel()
	reqIDHeader := "X-Request-ID"
	p := plugin{}
	cfg := make(map[string]any)
	cfg["Header"] = reqIDHeader
	cfg["Variant"] = vUUIDv7

	err := p.Setup(context.Background(), ika.InjectionContext{}, cfg)
	is.NoErr(err)

	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)

	// Test setting request ID
	r, err = p.ModifyRequest(r)
	is.NoErr(err)
	reqID := assertUUID(t, r.Header.Get(reqIDHeader))

	// Test not modified
	r, err = p.ModifyRequest(r)
	is.NoErr(err)
	is.Equal(reqID, assertUUID(t, r.Header.Get(reqIDHeader)))

	// Test appending
	p.cfg.Append = true
	r, err = p.ModifyRequest(r)
	is.NoErr(err)

	vals := r.Header.Values(reqIDHeader)
	reqID2 := assertUUID(t, vals[0])
	is.Equal(reqID, reqID2)
	reqID2 = assertUUID(t, vals[1])
	is.True(reqID != reqID2)
	is.True(len(vals) == 2)

	// Test overriding
	p.cfg.Override = true
	r, err = p.ModifyRequest(r)
	is.NoErr(err)
	reqID2 = assertUUID(t, r.Header.Get(reqIDHeader))
	is.True(reqID != reqID2)
	is.True(len(r.Header.Values(reqIDHeader)) == 1)
}

func assertUUID(t *testing.T, uuidStr string) uuid.UUID {
	t.Helper()
	is := is.New(t)
	uuid, err := uuid.Parse(uuidStr)
	is.NoErr(err)
	return uuid
}

func BenchmarkRand(b *testing.B) {
	b.Run("CryptoRand", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		buf := make([]byte, 16)
		for i := 0; i < b.N; i++ {
			_, err := rand.Read(buf)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("RandRand", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		buf := make([]byte, 16)
		for i := 0; i < b.N; i++ {
			_, err := rrand.Read(buf)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ChaChaRand", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		seed := [32]byte{}
		rand.Read(seed[:])
		cha := rrand2.NewChaCha8(seed)
		buf := make([]byte, 16)
		for i := 0; i < b.N; i++ {
			_, err := cha.Read(buf)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkIDGeneration(b *testing.B) {
	seed := [32]byte{}
	rand.Read(seed[:])
	cha := rrand2.NewChaCha8(seed)
	ksuid.SetRand(cha)
	uuid.SetRand(cha)
	uuid.EnableRandPool()

	b.Run("ksuid", func(b *testing.B) {
		_, err := ksuid.NewRandom()
		if err != nil {
			b.Fatal(err)
		}
	})

	b.Run("uuidv7", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := uuid.NewV7()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("uuidv4", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := uuid.NewRandom()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
