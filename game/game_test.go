package game

import (
	"strings"
	"testing"
)

func TestNewState(t *testing.T) {
	s := NewState()
	if s.IsInAir {
		t.Error("new state should not be in air")
	}
	if s.Score != 0 {
		t.Errorf("expected score 0, got %d", s.Score)
	}
	if s.GameOver {
		t.Error("new state should not be game over")
	}
	if len(s.Obstacles) != 0 {
		t.Errorf("expected no obstacles, got %d", len(s.Obstacles))
	}
}

func TestJump(t *testing.T) {
	tests := []struct {
		name       string
		initial    State
		wantInAir  bool
		wantJumpTime float64
	}{
		{
			name:       "jump from ground",
			initial:    State{IsInAir: false},
			wantInAir:  true,
			wantJumpTime: JumpDuration,
		},
		{
			name:       "cannot jump while in air",
			initial:    State{IsInAir: true, JumpTimeLeft: 0.2},
			wantInAir:  true,
			wantJumpTime: 0.2,
		},
		{
			name:       "cannot jump when game over",
			initial:    State{GameOver: true},
			wantInAir:  false,
			wantJumpTime: 0,
		},
		{
			name:       "cannot jump when paused",
			initial:    State{IsPaused: true},
			wantInAir:  false,
			wantJumpTime: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Jump(tt.initial)
			if result.IsInAir != tt.wantInAir {
				t.Errorf("IsInAir: got %v, want %v", result.IsInAir, tt.wantInAir)
			}
			if result.JumpTimeLeft != tt.wantJumpTime {
				t.Errorf("JumpTimeLeft: got %v, want %v", result.JumpTimeLeft, tt.wantJumpTime)
			}
		})
	}
}

func TestUpdateJump(t *testing.T) {
	tests := []struct {
		name         string
		initial      State
		dt           float64
		wantInAir    bool
		wantTimeLeft float64
	}{
		{
			name:         "decrement jump time",
			initial:      State{IsInAir: true, JumpTimeLeft: 0.3},
			dt:           0.1,
			wantInAir:    true,
			wantTimeLeft: 0.2,
		},
		{
			name:         "land when time expires",
			initial:      State{IsInAir: true, JumpTimeLeft: 0.05},
			dt:           0.1,
			wantInAir:    false,
			wantTimeLeft: 0,
		},
		{
			name:         "no change when on ground",
			initial:      State{IsInAir: false},
			dt:           0.1,
			wantInAir:    false,
			wantTimeLeft: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := updateJump(tt.initial, tt.dt)
			if result.IsInAir != tt.wantInAir {
				t.Errorf("IsInAir: got %v, want %v", result.IsInAir, tt.wantInAir)
			}
			if result.JumpTimeLeft < tt.wantTimeLeft-0.001 || result.JumpTimeLeft > tt.wantTimeLeft+0.001 {
				t.Errorf("JumpTimeLeft: got %v, want %v", result.JumpTimeLeft, tt.wantTimeLeft)
			}
		})
	}
}

func TestUpdateObstacles(t *testing.T) {
	s := State{
		Obstacles:   []float64{50, 30, 10, -1},
		ElapsedTime: 0,
	}
	result := updateObstacles(s, 0.1)

	if len(result.Obstacles) != 3 {
		t.Errorf("expected 3 obstacles after removal, got %d", len(result.Obstacles))
	}

	for i, x := range result.Obstacles {
		if x >= s.Obstacles[i] {
			t.Errorf("obstacle %d did not move left: was %v, now %v", i, s.Obstacles[i], x)
		}
	}
}

func TestSpawnObstacle(t *testing.T) {
	width := 100

	t.Run("spawn when interval elapsed", func(t *testing.T) {
		s := State{TimeSinceSpawn: BaseSpawnInterval + 0.1}
		result := spawnObstacle(s, width)
		if len(result.Obstacles) != 1 {
			t.Errorf("expected 1 obstacle, got %d", len(result.Obstacles))
		}
		if result.Obstacles[0] != float64(width) {
			t.Errorf("expected obstacle at %d, got %v", width, result.Obstacles[0])
		}
		if result.TimeSinceSpawn != 0 {
			t.Errorf("TimeSinceSpawn should reset, got %v", result.TimeSinceSpawn)
		}
	})

	t.Run("no spawn when interval not elapsed", func(t *testing.T) {
		s := State{TimeSinceSpawn: 0.5}
		result := spawnObstacle(s, width)
		if len(result.Obstacles) != 0 {
			t.Errorf("expected no obstacles, got %d", len(result.Obstacles))
		}
	})
}

func TestCheckCollision(t *testing.T) {
	tests := []struct {
		name      string
		state     State
		wantHit   bool
	}{
		{
			name:    "no collision when no obstacles",
			state:   State{Obstacles: []float64{}},
			wantHit: false,
		},
		{
			name:    "no collision when obstacle far away",
			state:   State{Obstacles: []float64{50}},
			wantHit: false,
		},
		{
			name:    "collision when obstacle at dino position",
			state:   State{Obstacles: []float64{1}},
			wantHit: true,
		},
		{
			name:    "collision at hitbox boundary",
			state:   State{Obstacles: []float64{DinoHitboxEnd}},
			wantHit: true,
		},
		{
			name:    "no collision when jumping",
			state:   State{IsInAir: true, Obstacles: []float64{1}},
			wantHit: false,
		},
		{
			name:    "no collision when obstacle just passed",
			state:   State{Obstacles: []float64{-0.1}},
			wantHit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkCollision(tt.state)
			if result != tt.wantHit {
				t.Errorf("got %v, want %v", result, tt.wantHit)
			}
		})
	}
}

func TestTick(t *testing.T) {
	t.Run("does not update when paused", func(t *testing.T) {
		s := State{IsPaused: true, Score: 10}
		result := Tick(s, 0.1, 100)
		if result.Score != 10 {
			t.Errorf("score should not change when paused, got %d", result.Score)
		}
	})

	t.Run("does not update when game over", func(t *testing.T) {
		s := State{GameOver: true, Score: 10}
		result := Tick(s, 0.1, 100)
		if result.Score != 10 {
			t.Errorf("score should not change when game over, got %d", result.Score)
		}
	})

	t.Run("increments score", func(t *testing.T) {
		s := NewState()
		result := Tick(s, 0.016, 100)
		if result.Score != 1 {
			t.Errorf("expected score 1, got %d", result.Score)
		}
	})

	t.Run("updates elapsed time", func(t *testing.T) {
		s := NewState()
		result := Tick(s, 0.1, 100)
		if result.ElapsedTime < 0.099 || result.ElapsedTime > 0.101 {
			t.Errorf("expected elapsed time ~0.1, got %v", result.ElapsedTime)
		}
	})

	t.Run("sets game over on collision", func(t *testing.T) {
		s := State{Obstacles: []float64{1}}
		result := Tick(s, 0.016, 100)
		if !result.GameOver {
			t.Error("expected game over on collision")
		}
	})

	t.Run("updates high score on game over", func(t *testing.T) {
		s := State{Score: 100, HighScore: 50, Obstacles: []float64{1}}
		result := Tick(s, 0.016, 100)
		if result.HighScore != 100 {
			t.Errorf("expected high score 100, got %d", result.HighScore)
		}
	})
}

func TestRestart(t *testing.T) {
	s := State{
		Score:     500,
		HighScore: 1000,
		GameOver:  true,
		Obstacles: []float64{10, 20},
	}
	result := Restart(s)

	if result.Score != 0 {
		t.Errorf("score should reset, got %d", result.Score)
	}
	if result.HighScore != 1000 {
		t.Errorf("high score should preserve, got %d", result.HighScore)
	}
	if result.GameOver {
		t.Error("game over should reset")
	}
	if len(result.Obstacles) != 0 {
		t.Errorf("obstacles should clear, got %d", len(result.Obstacles))
	}
}

func TestTogglePause(t *testing.T) {
	t.Run("pause when playing", func(t *testing.T) {
		s := State{IsPaused: false}
		result := TogglePause(s)
		if !result.IsPaused {
			t.Error("should be paused")
		}
	})

	t.Run("unpause when paused", func(t *testing.T) {
		s := State{IsPaused: true}
		result := TogglePause(s)
		if result.IsPaused {
			t.Error("should be unpaused")
		}
	})

	t.Run("cannot toggle when game over", func(t *testing.T) {
		s := State{GameOver: true, IsPaused: false}
		result := TogglePause(s)
		if result.IsPaused {
			t.Error("should not toggle when game over")
		}
	})
}

func TestCurrentSpeed(t *testing.T) {
	speed0 := currentSpeed(0)
	if speed0 != BaseObstacleSpeed {
		t.Errorf("expected base speed %v at time 0, got %v", BaseObstacleSpeed, speed0)
	}

	speed10 := currentSpeed(10)
	expectedSpeed10 := BaseObstacleSpeed + SpeedIncreaseRate
	if speed10 < expectedSpeed10-0.001 || speed10 > expectedSpeed10+0.001 {
		t.Errorf("expected speed %v at time 10, got %v", expectedSpeed10, speed10)
	}
}

func TestSpawnInterval(t *testing.T) {
	interval0 := spawnInterval(0)
	if interval0 != BaseSpawnInterval {
		t.Errorf("expected base interval %v at score 0, got %v", BaseSpawnInterval, interval0)
	}

	intervalHigh := spawnInterval(10000)
	if intervalHigh < MinSpawnInterval {
		t.Errorf("interval should not go below min %v, got %v", MinSpawnInterval, intervalHigh)
	}
}

func TestRender(t *testing.T) {
	t.Run("minimum width fallback", func(t *testing.T) {
		sky, ground := Render(State{}, 10)
		if sky != CloudEmoji {
			t.Errorf("expected cloud emoji for narrow width, got %q", sky)
		}
		if ground != DinoEmoji {
			t.Errorf("expected dino emoji for narrow width, got %q", ground)
		}
	})

	t.Run("dino on ground when not jumping", func(t *testing.T) {
		_, ground := Render(State{IsInAir: false}, 80)
		if !strings.HasPrefix(ground, DinoEmoji) {
			t.Errorf("expected dino at start of ground line, got %q", ground[:10])
		}
	})

	t.Run("dino in sky when jumping", func(t *testing.T) {
		sky, ground := Render(State{IsInAir: true}, 80)
		if !strings.HasPrefix(sky, DinoEmoji) {
			t.Errorf("expected dino at start of sky line, got %q", sky[:10])
		}
		if strings.HasPrefix(ground, DinoEmoji) {
			t.Error("dino should not be on ground when jumping")
		}
	})

	t.Run("dead emoji when game over", func(t *testing.T) {
		_, ground := Render(State{GameOver: true}, 80)
		if !strings.HasPrefix(ground, DeadEmoji) {
			t.Errorf("expected dead emoji, got %q", ground[:10])
		}
	})

	t.Run("score displayed", func(t *testing.T) {
		_, ground := Render(State{Score: 42, HighScore: 100}, 80)
		if !strings.Contains(ground, "Score: 00042") {
			t.Errorf("expected score in ground line, got %q", ground)
		}
		if !strings.Contains(ground, "HI: 00100") {
			t.Errorf("expected high score in ground line, got %q", ground)
		}
	})

	t.Run("restart hint when game over", func(t *testing.T) {
		_, ground := Render(State{GameOver: true}, 80)
		if !strings.Contains(ground, "[R]estart") {
			t.Errorf("expected restart hint, got %q", ground)
		}
	})

	t.Run("obstacles rendered", func(t *testing.T) {
		_, ground := Render(State{Obstacles: []float64{20}}, 80)
		if !strings.Contains(ground, ObstacleEmoji) {
			t.Errorf("expected obstacle in ground line, got %q", ground)
		}
	})
}

func TestFormatCountdown(t *testing.T) {
	t.Run("countdown centered", func(t *testing.T) {
		sky := strings.Repeat(" ", 80)
		result := FormatCountdown(sky, 3, 80)
		if !strings.Contains(result, "▶ 3 ◀") {
			t.Errorf("expected countdown in result, got %q", result)
		}
	})

	t.Run("no change when countdown zero", func(t *testing.T) {
		sky := "test sky"
		result := FormatCountdown(sky, 0, 80)
		if result != sky {
			t.Errorf("expected unchanged sky, got %q", result)
		}
	})
}
