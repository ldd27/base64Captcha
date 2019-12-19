// Copyright 2017 Eric Zhou. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package base64Captcha

import (
	"encoding/binary"
	"math/rand"
)

// siprng is PRNG based on SipHash-2-4.
// (Note: it's not safe to use a single siprng from multiple goroutines.)
type siprng struct {
	k0, k1, ctr uint64
}

// siphash implements SipHash-2-4, accepting a uint64 as a message.
func siphash(k0, k1, m uint64) uint64 {
	// Initialization.
	v0 := k0 ^ 0x736f6d6570736575
	v1 := k1 ^ 0x646f72616e646f6d
	v2 := k0 ^ 0x6c7967656e657261
	v3 := k1 ^ 0x7465646279746573
	t := uint64(8) << 56

	// Compression.
	v3 ^= m

	// Round 1.
	v0 += v1
	v1 = v1<<13 | v1>>(64-13)
	v1 ^= v0
	v0 = v0<<32 | v0>>(64-32)

	v2 += v3
	v3 = v3<<16 | v3>>(64-16)
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>(64-21)
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>(64-17)
	v1 ^= v2
	v2 = v2<<32 | v2>>(64-32)

	// Round 2.
	v0 += v1
	v1 = v1<<13 | v1>>(64-13)
	v1 ^= v0
	v0 = v0<<32 | v0>>(64-32)

	v2 += v3
	v3 = v3<<16 | v3>>(64-16)
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>(64-21)
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>(64-17)
	v1 ^= v2
	v2 = v2<<32 | v2>>(64-32)

	v0 ^= m

	// Compress last block.
	v3 ^= t

	// Round 1.
	v0 += v1
	v1 = v1<<13 | v1>>(64-13)
	v1 ^= v0
	v0 = v0<<32 | v0>>(64-32)

	v2 += v3
	v3 = v3<<16 | v3>>(64-16)
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>(64-21)
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>(64-17)
	v1 ^= v2
	v2 = v2<<32 | v2>>(64-32)

	// Round 2.
	v0 += v1
	v1 = v1<<13 | v1>>(64-13)
	v1 ^= v0
	v0 = v0<<32 | v0>>(64-32)

	v2 += v3
	v3 = v3<<16 | v3>>(64-16)
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>(64-21)
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>(64-17)
	v1 ^= v2
	v2 = v2<<32 | v2>>(64-32)

	v0 ^= t

	// Finalization.
	v2 ^= 0xff

	// Round 1.
	v0 += v1
	v1 = v1<<13 | v1>>(64-13)
	v1 ^= v0
	v0 = v0<<32 | v0>>(64-32)

	v2 += v3
	v3 = v3<<16 | v3>>(64-16)
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>(64-21)
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>(64-17)
	v1 ^= v2
	v2 = v2<<32 | v2>>(64-32)

	// Round 2.
	v0 += v1
	v1 = v1<<13 | v1>>(64-13)
	v1 ^= v0
	v0 = v0<<32 | v0>>(64-32)

	v2 += v3
	v3 = v3<<16 | v3>>(64-16)
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>(64-21)
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>(64-17)
	v1 ^= v2
	v2 = v2<<32 | v2>>(64-32)

	// Round 3.
	v0 += v1
	v1 = v1<<13 | v1>>(64-13)
	v1 ^= v0
	v0 = v0<<32 | v0>>(64-32)

	v2 += v3
	v3 = v3<<16 | v3>>(64-16)
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>(64-21)
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>(64-17)
	v1 ^= v2
	v2 = v2<<32 | v2>>(64-32)

	// Round 4.
	v0 += v1
	v1 = v1<<13 | v1>>(64-13)
	v1 ^= v0
	v0 = v0<<32 | v0>>(64-32)

	v2 += v3
	v3 = v3<<16 | v3>>(64-16)
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>(64-21)
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>(64-17)
	v1 ^= v2
	v2 = v2<<32 | v2>>(64-32)

	return v0 ^ v1 ^ v2 ^ v3
}

// Seed sets a new secret seed for PRNG.
func (p *siprng) Seed(k [16]byte) {
	p.k0 = binary.LittleEndian.Uint64(k[0:8])
	p.k1 = binary.LittleEndian.Uint64(k[8:16])
	p.ctr = 1
}

// Uint64 returns a new pseudorandom uint64.
func (p *siprng) Uint64() uint64 {
	v := siphash(p.k0, p.k1, p.ctr)
	p.ctr++
	return v
}

func (p *siprng) Bytes(n int) []byte {
	// Since we don't have a buffer for generated bytes in siprng state,
	// we just generate enough 8-byte blocks and then cut the result to the
	// required length. Doing it this way, we lose generated bytes, and we
	// don't get the strictly sequential deterministic output from PRNG:
	// calling Uint64() and then Bytes(3) produces different output than
	// when calling them in the reverse order, but for our applications
	// this is OK.
	numBlocks := (n + 8 - 1) / 8
	b := make([]byte, numBlocks*8)
	for i := 0; i < len(b); i += 8 {

		binary.LittleEndian.PutUint64(b[i:], rand.Uint64())
	}
	return b[:n]
}

func (p *siprng) Int63() int64 {
	return int64(p.Uint64() & 0x7fffffffffffffff)
}

func (p *siprng) Uint32() uint32 {
	return uint32(p.Uint64())
}

func (p *siprng) Int31() int32 {
	return int32(p.Uint32() & 0x7fffffff)
}

func (p *siprng) Intn(n int) int {
	if n <= 0 {
		panic("invalid argument to Intn")
	}
	if n <= 1<<31-1 {
		return int(p.Int31n(int32(n)))
	}
	return int(p.Int63n(int64(n)))
}

func (p *siprng) Int63n(n int64) int64 {
	if n <= 0 {
		panic("invalid argument to Int63n")
	}
	max := int64((1 << 63) - 1 - (1<<63)%uint64(n))
	v := p.Int63()
	for v > max {
		v = p.Int63()
	}
	return v % n
}

func (p *siprng) Int31n(n int32) int32 {
	if n <= 0 {
		panic("invalid argument to Int31n")
	}
	max := int32((1 << 31) - 1 - (1<<31)%uint32(n))
	v := p.Int31()
	for v > max {
		v = p.Int31()
	}
	return v % n
}

func (p *siprng) Float64() float64 { return float64(p.Int63()) / (1 << 63) }

// Int returns a pseudorandom int in range [from, to].
func (p *siprng) Int(from, to int) int {
	return p.Intn(to+1-from) + from
}

// Float returns a pseudorandom float64 in range [from, to].
func (p *siprng) Float(from, to float64) float64 {
	return (to-from)*p.Float64() + from
}
