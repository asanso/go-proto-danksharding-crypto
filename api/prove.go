package api

import (
	"github.com/crate-crypto/go-proto-danksharding-crypto/internal/kzg"
	"github.com/crate-crypto/go-proto-danksharding-crypto/serialization"
)

// spec: blob_to_kzg_commitments
// For now we call the method that calls multiple Blobs as a sub-routine
func (c *Context) BlobToCommitment(blob serialization.Blob) (serialization.Commitment, error) {
	commitments, err := c.BlobsToCommitments([]serialization.Blob{blob})
	if err != nil {
		return serialization.Commitment{}, err
	}
	return commitments[0], nil
}
func (c *Context) BlobsToCommitments(blobs []serialization.Blob) (serialization.Commitments, error) {
	// Deserialization
	//
	// 1. Deserialize the Blobs into polynomial objects
	polys, err := serialization.DeserializeBlobs(blobs)
	if err != nil {
		return nil, err
	}

	// 2. Commit to polynomials
	comms, err := kzg.CommitToPolynomials(polys, c.commitKey)
	if err != nil {
		return nil, err
	}

	// Serialization
	//
	// 3. Serialize commitments
	serComms := serialization.SerializeG1Points(comms)

	return serComms, nil
}

// Note: This method does not check that the commitment corresponds
// to the `blob`.
// The method does still check that the commitment is a valid commitment.
// One should check this externally or call `BlobToCommitment`
func (c *Context) ComputeBlobKZGProof(blob serialization.Blob, serializedComm serialization.Commitment) (serialization.KZGProof, serialization.Scalar, error) {
	// Deserialization
	//
	// 1. Deserialize the `Blob`  into a polynomial
	//
	poly, err := serialization.DeserializeBlob(blob)
	if err != nil {
		return serialization.KZGProof{}, [32]byte{}, err
	}
	// Deserialize the commitment -- we only do this to check
	// if it is in the correct subgroup
	_, err = serialization.DeserializeG1Point(serializedComm)
	if err != nil {
		return serialization.KZGProof{}, [32]byte{}, err
	}

	// 2. Compute Fiat-Shamir challenge
	evaluationChallenge := computeChallenge(blob, serializedComm)

	// 3. Create opening proof
	openingProof, err := kzg.Open(c.domain, poly, evaluationChallenge, c.commitKey)
	if err != nil {
		return serialization.KZGProof{}, [32]byte{}, err
	}

	// Serialization
	//
	// 4. Serialize values
	//
	// Polynomial commitment
	//
	// Quotient commitment
	serProof := serialization.SerializeG1Point(openingProof.QuotientComm)
	//
	// Claimed value -- Reverse it to use little endian
	claimedValueBytes := serialization.SerializeScalar(openingProof.ClaimedValue)

	return serProof, claimedValueBytes, nil
}

func (c *Context) ComputeKZGProof(blob serialization.Blob, inputPointBytes serialization.Scalar) (serialization.KZGProof, serialization.Scalar, error) {
	// Deserialization
	//
	// 1. Deserialize the `Blob` into a polynomial
	//
	poly, err := serialization.DeserializeBlob(blob)
	if err != nil {
		return serialization.KZGProof{}, [32]byte{}, err
	}

	// 2. Deserialize input point
	inputPoint, err := serialization.DeserializeScalar(inputPointBytes)
	if err != nil {
		return serialization.KZGProof{}, [32]byte{}, err
	}

	// 3. Create opening proof
	openingProof, err := kzg.Open(c.domain, poly, inputPoint, c.commitKey)
	if err != nil {
		return serialization.KZGProof{}, [32]byte{}, err
	}

	// Serialization
	//
	// 4. Serialize values
	//

	//
	// Quotient commitment
	serProof := serialization.SerializeG1Point(openingProof.QuotientComm)
	//
	// Claimed value
	claimedValueBytes := serialization.SerializeScalar(openingProof.ClaimedValue)

	return serProof, claimedValueBytes, nil
}
