package factorio

// Position is a map/world coordinate. Factorio uses floating-point tile
// coordinates with +x east and +y south.
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}
