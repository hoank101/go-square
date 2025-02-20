package inclusion_test

import (
	"bytes"
	"crypto/sha256"
	"testing"

	"github.com/celestiaorg/go-square/blob"
	"github.com/celestiaorg/go-square/inclusion"
	"github.com/celestiaorg/go-square/namespace"
	"github.com/celestiaorg/go-square/shares"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_MerkleMountainRangeHeights(t *testing.T) {
	type test struct {
		totalSize  uint64
		squareSize uint64
		expected   []uint64
	}
	tests := []test{
		{
			totalSize:  11,
			squareSize: 4,
			expected:   []uint64{4, 4, 2, 1},
		},
		{
			totalSize:  2,
			squareSize: 64,
			expected:   []uint64{2},
		},
		{
			totalSize:  64,
			squareSize: 8,
			expected:   []uint64{8, 8, 8, 8, 8, 8, 8, 8},
		},
		// Height
		// 3              x                               x
		//              /    \                         /    \
		//             /      \                       /      \
		//            /        \                     /        \
		//           /          \                   /          \
		// 2        x            x                 x            x
		//        /   \        /   \             /   \        /   \
		// 1     x     x      x     x           x     x      x     x         x
		//      / \   / \    / \   / \         / \   / \    / \   / \      /   \
		// 0   0   1 2   3  4   5 6   7       8   9 10  11 12 13 14  15   16   17    18
		{
			totalSize:  19,
			squareSize: 8,
			expected:   []uint64{8, 8, 2, 1},
		},
	}
	for _, tt := range tests {
		res, err := inclusion.MerkleMountainRangeSizes(tt.totalSize, tt.squareSize)
		require.NoError(t, err)
		assert.Equal(t, tt.expected, res)
	}
}

// TestCreateCommitment will fail if a change is made to share encoding or how
// the commitment is calculated. If this is the case, the expected commitment
// bytes will need to be updated.
func TestCreateCommitment(t *testing.T) {
	ns1 := namespace.MustNewV0(bytes.Repeat([]byte{0x1}, namespace.NamespaceVersionZeroIDSize))

	type test struct {
		name         string
		namespace    namespace.Namespace
		blob         []byte
		expected     []byte
		expectErr    bool
		shareVersion uint8
	}
	tests := []test{
		{
			name:         "blob of 2 shares succeeds",
			namespace:    ns1,
			blob:         bytes.Repeat([]byte{0xFF}, shares.AvailableBytesFromSparseShares(2)),
			expected:     []byte{0x31, 0xf5, 0x15, 0x6d, 0x5d, 0xb9, 0xa7, 0xf5, 0xb4, 0x3b, 0x29, 0x7a, 0x14, 0xc0, 0x70, 0xc2, 0xcc, 0x4e, 0xf3, 0xd6, 0x9d, 0x87, 0xed, 0x8, 0xad, 0xdd, 0x21, 0x6d, 0x9b, 0x9f, 0xa1, 0x18},
			shareVersion: shares.ShareVersionZero,
		},
		{
			name:         "blob with unsupported share version should return error",
			namespace:    ns1,
			blob:         bytes.Repeat([]byte{0xFF}, shares.AvailableBytesFromSparseShares(2)),
			expectErr:    true,
			shareVersion: uint8(1), // unsupported share version
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blob := &blob.Blob{
				NamespaceId:      tt.namespace.ID,
				Data:             tt.blob,
				ShareVersion:     uint32(tt.shareVersion),
				NamespaceVersion: uint32(tt.namespace.Version),
			}
			res, err := inclusion.CreateCommitment(blob, twoLeafMerkleRoot, defaultSubtreeRootThreshold)
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, res)
		})
	}
}

func twoLeafMerkleRoot(data [][]byte) []byte {
	if len(data) != 2 {
		panic("data must have exactly 2 elements")
	}
	h1 := sha256.Sum256(data[0])
	h2 := sha256.Sum256(data[1])
	sum := sha256.Sum256(append(h1[:], h2[:]...))
	return sum[:]
}
