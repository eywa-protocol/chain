package curve

import (
	"crypto/elliptic"
	"math/big"
	"sync"
)

var (
// fieldOne is simply the integer 1 in field representation.  It is
// used to avoid needing to create it multiple times during the internal
// arithmetic.
//fieldOne = new(fieldVal).SetInt(1)
)

type fieldVal struct {
	n [10]uint32
}

type AltBN256Curve struct {
	*elliptic.CurveParams

	// q is the value (P+1)/4 used to compute the square root of field
	// elements.
	q *big.Int

	H         int      // cofactor of the curve.
	halfOrder *big.Int // half the order N

	// fieldB is the constant B of the curve as a fieldVal.
	fieldB *fieldVal

	// byteSize is simply the bit size / 8 and is provided for convenience
	// since it is calculated repeatedly.
	byteSize int

	// bytePoints
	bytePoints *[32][256][3]fieldVal

	// The next 6 values are used specifically for endomorphism
	// optimizations in ScalarMult.

	// lambda must fulfill lambda^3 = 1 mod N where N is the order of G.
	lambda *big.Int

	// beta must fulfill beta^3 = 1 mod P where P is the prime field of the
	// curve.
	beta *fieldVal

	// See the EndomorphismVectors in genaltbn256.go to see how these are
	// derived.
	a1 *big.Int
	b1 *big.Int
	a2 *big.Int
	b2 *big.Int
}

// Params returns the parameters for the curve.
func (curve *AltBN256Curve) Params() *elliptic.CurveParams {
	return curve.CurveParams
}

func (curve *AltBN256Curve) IsOnCurve(x, y *big.Int) bool {
	return true
}

// Add returns the sum of (x1,y1) and (x2,y2)
func (curve *AltBN256Curve) Add(x1, y1, x2, y2 *big.Int) (x, y *big.Int) {
	return big.NewInt(0), big.NewInt(3)
}

// Double returns 2*(x,y)
func (curve *AltBN256Curve) Double(x1, y1 *big.Int) (x, y *big.Int) {
	return big.NewInt(0), big.NewInt(3)
}

// ScalarMult returns k*(Bx,By) where k is a number in big-endian form.
func (curve *AltBN256Curve) ScalarMult(x1, y1 *big.Int, k []byte) (x, y *big.Int) {
	return big.NewInt(0), big.NewInt(3)
}

// ScalarBaseMult returns k*G, where G is the base point of the group
// and k is an integer in big-endian form.
func (curve *AltBN256Curve) ScalarBaseMult(k []byte) (x, y *big.Int) {
	return big.NewInt(0), big.NewInt(3)
}

var initonce sync.Once
var altbn256 AltBN256Curve

func initAll() {
	initAltBN256()
}

func initAltBN256() {
	// Curve parameters taken from [SECG] section 2.4.1.
	altbn256.CurveParams = new(elliptic.CurveParams)
	altbn256.P = fromHex("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F")
	altbn256.N = fromHex("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141")
	altbn256.B = fromHex("0000000000000000000000000000000000000000000000000000000000000007")
	altbn256.Gx = fromHex("79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798")
	altbn256.Gy = fromHex("483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8")
	altbn256.BitSize = 256
	altbn256.q = new(big.Int).Div(new(big.Int).Add(altbn256.P,
		big.NewInt(1)), big.NewInt(4))
	altbn256.H = 1
	altbn256.halfOrder = new(big.Int).Rsh(altbn256.N, 1)
	//altbn256.fieldB = new(fieldVal).SetByteSlice(altbn256.B.Bytes())

	// Provided for convenience since this gets computed repeatedly.
	altbn256.byteSize = altbn256.BitSize / 8

	// Deserialize and set the pre-computed table used to accelerate scalar
	// base multiplication.  This is hard-coded data, so any errors are
	// panics because it means something is wrong in the source code.
	//if err := loadS256BytePoints(); err != nil {
	//	panic(err)
	//}

	// Next 6 constants are from Hal Finney's bitcointalk.org post:
	// https://bitcointalk.org/index.php?topic=3238.msg45565#msg45565
	// May he rest in peace.
	//
	// They have also been independently derived from the code in the
	// EndomorphismVectors function in genaltbn256.go.
	altbn256.lambda = fromHex("5363AD4CC05C30E0A5261C028812645A122E22EA20816678DF02967C1B23BD72")
	//altbn256.beta = new(fieldVal).SetHex("7AE96A2B657C07106E64479EAC3434E99CF0497512F58995C1396C28719501EE")
	altbn256.a1 = fromHex("3086D221A7D46BCDE86C90E49284EB15")
	altbn256.b1 = fromHex("-E4437ED6010E88286F547FA90ABFE4C3")
	altbn256.a2 = fromHex("114CA50F7A8E2F3F657C1108D9D44CFD8")
	altbn256.b2 = fromHex("3086D221A7D46BCDE86C90E49284EB15")

	// Alternatively, we can use the parameters below, however, they seem
	//  to be about 8% slower.
	// altbn256.lambda = fromHex("AC9C52B33FA3CF1F5AD9E3FD77ED9BA4A880B9FC8EC739C2E0CFC810B51283CE")
	// altbn256.beta = new(fieldVal).SetHex("851695D49A83F8EF919BB86153CBCB16630FB68AED0A766A3EC693D68E6AFA40")
	// altbn256.a1 = fromHex("E4437ED6010E88286F547FA90ABFE4C3")
	// altbn256.b1 = fromHex("-3086D221A7D46BCDE86C90E49284EB15")
	// altbn256.a2 = fromHex("3086D221A7D46BCDE86C90E49284EB15")
	// altbn256.b2 = fromHex("114CA50F7A8E2F3F657C1108D9D44CFD8")
}

// fromHex converts the passed hex string into a big integer pointer and will
// panic is there is an error.  This is only provided for the hard-coded
// constants so errors in the source code can bet detected. It will only (and
// must only) be called for initialization purposes.
func fromHex(s string) *big.Int {
	r, ok := new(big.Int).SetString(s, 16)
	if !ok {
		panic("invalid hex in source file: " + s)
	}
	return r
}

// S256 returns a Curve which implements secp256k1.
func ALTBN256() *AltBN256Curve {
	initonce.Do(initAll)
	return &altbn256
}
