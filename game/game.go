package game

const (
	JumpDuration      = 0.4
	BaseObstacleSpeed = 40.0
	SpeedIncreaseRate = 2.0
	BaseSpawnInterval = 2.0
	MinSpawnInterval  = 0.8
	SpawnDecreaseRate = 0.1
	DinoHitboxEnd     = 2.0
)

type State struct {
	IsInAir        bool
	JumpTimeLeft   float64
	Obstacles      []float64
	Score          int
	HighScore      int
	GameOver       bool
	IsPaused       bool
	ElapsedTime    float64
	TimeSinceSpawn float64
}

func NewState() State {
	return State{}
}

func Tick(s State, dt float64, width int) State {
	if s.IsPaused || s.GameOver {
		return s
	}

	s.ElapsedTime += dt
	s.TimeSinceSpawn += dt

	s = updateJump(s, dt)
	s = updateObstacles(s, dt)
	s = spawnObstacle(s, width)

	if checkCollision(s) {
		s.GameOver = true
		if s.Score > s.HighScore {
			s.HighScore = s.Score
		}
		return s
	}

	s.Score++
	if s.Score > s.HighScore {
		s.HighScore = s.Score
	}

	return s
}

func Jump(s State) State {
	if s.IsInAir || s.GameOver || s.IsPaused {
		return s
	}
	s.IsInAir = true
	s.JumpTimeLeft = JumpDuration
	return s
}

func Restart(s State) State {
	return State{
		HighScore: s.HighScore,
	}
}

func TogglePause(s State) State {
	if s.GameOver {
		return s
	}
	s.IsPaused = !s.IsPaused
	return s
}

func updateJump(s State, dt float64) State {
	if !s.IsInAir {
		return s
	}
	s.JumpTimeLeft -= dt
	if s.JumpTimeLeft <= 0 {
		s.IsInAir = false
		s.JumpTimeLeft = 0
	}
	return s
}

func updateObstacles(s State, dt float64) State {
	speed := currentSpeed(s.ElapsedTime)
	newObstacles := make([]float64, 0, len(s.Obstacles))
	for _, x := range s.Obstacles {
		newX := x - speed*dt
		if newX > -2 {
			newObstacles = append(newObstacles, newX)
		}
	}
	s.Obstacles = newObstacles
	return s
}

func spawnObstacle(s State, width int) State {
	interval := spawnInterval(s.Score)
	if s.TimeSinceSpawn < interval {
		return s
	}
	s.TimeSinceSpawn = 0
	s.Obstacles = append(s.Obstacles, float64(width))
	return s
}

func checkCollision(s State) bool {
	if s.IsInAir {
		return false
	}
	for _, x := range s.Obstacles {
		if x >= 0 && x <= DinoHitboxEnd {
			return true
		}
	}
	return false
}

func currentSpeed(elapsed float64) float64 {
	return BaseObstacleSpeed + (elapsed/10.0)*SpeedIncreaseRate
}

func spawnInterval(score int) float64 {
	interval := BaseSpawnInterval - (float64(score)/100.0)*SpawnDecreaseRate
	if interval < MinSpawnInterval {
		return MinSpawnInterval
	}
	return interval
}
