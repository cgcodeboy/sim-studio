package scenario

import "testing"

func TestComplexJuicePlant(t *testing.T) { play(t, complexScenes()[0]) }
func TestComplexPartLine(t *testing.T)   { play(t, complexScenes()[1]) }

func complexScenes() []scene {
	return []scene{juiceCampusScene(), partsCampusScene()}
}
