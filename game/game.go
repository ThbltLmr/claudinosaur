package game

const (
	JumpDuration       = 0.4
	BaseObstacleSpeed  = 40.0
	SpeedIncreaseRate  = 2.0
	BaseSpawnInterval  = 2.0
	MinSpawnInterval   = 0.8
	SpawnDecreaseRate  = 0.1
	DinoHitboxEnd      = 2.0
	GameOverDelay      = 2.0

	CloudSpeed      = 15.0
	MinCloudSpacing = 10.0

	BirdScoreThreshold = 200
	BirdSpawnInterval  = 3.0
	BirdSafeZone       = 5.0
)

type State struct {
	IsInAir          bool
	JumpTimeLeft     float64
	Obstacles        []float64
	Score            int
	HighScore        int
	GameOver         bool
	GameOverTime     float64
	IsPaused         bool
	ElapsedTime      float64
	TimeSinceSpawn   float64
	Clouds           []float64
	Birds            []float64
	TimeSinceBird    float64
}

func NewState() State {
	return State{
		Clouds: initialClouds(),
	}
}

func initialClouds() []float64 {
	return []float64{4, 20, 45}
}

func Tick(s State, dt float64, width int) State {
	if s.IsPaused {
		return s
	}

	if s.GameOver {
		s.GameOverTime += dt
		if s.GameOverTime >= GameOverDelay {
			return Restart(s)
		}
		return s
	}

	s.ElapsedTime += dt
	s.TimeSinceSpawn += dt

	s = updateJump(s, dt)
	s = updateObstacles(s, dt)
	s = updateClouds(s, dt)
	s = updateBirds(s, dt)
	s = spawnObstacle(s, width)
	s = spawnCloud(s, width)
	s = spawnBird(s, width)

	if checkCollision(s) || checkBirdCollision(s) {
		s.GameOver = true
		s.GameOverTime = 0
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
		Clouds:    initialClouds(),
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

func updateClouds(s State, dt float64) State {
	newClouds := make([]float64, 0, len(s.Clouds))
	for _, x := range s.Clouds {
		newX := x - CloudSpeed*dt
		if newX > -2 {
			newClouds = append(newClouds, newX)
		}
	}
	s.Clouds = newClouds
	return s
}

func spawnCloud(s State, width int) State {
	if len(s.Clouds) == 0 {
		s.Clouds = append(s.Clouds, float64(width))
		return s
	}
	rightmost := s.Clouds[len(s.Clouds)-1]
	if float64(width)-rightmost >= MinCloudSpacing {
		s.Clouds = append(s.Clouds, float64(width))
	}
	return s
}

func updateBirds(s State, dt float64) State {
	s.TimeSinceBird += dt
	speed := currentSpeed(s.ElapsedTime)
	newBirds := make([]float64, 0, len(s.Birds))
	for _, x := range s.Birds {
		newX := x - speed*dt
		if newX > -2 {
			newBirds = append(newBirds, newX)
		}
	}
	s.Birds = newBirds
	return s
}

func spawnBird(s State, width int) State {
	if s.Score < BirdScoreThreshold {
		return s
	}
	if s.TimeSinceBird < BirdSpawnInterval {
		return s
	}
	spawnX := float64(width)
	if hasConflict(s.Clouds, spawnX, BirdSafeZone) || hasConflict(s.Obstacles, spawnX, BirdSafeZone) {
		return s
	}
	s.TimeSinceBird = 0
	s.Birds = append(s.Birds, spawnX)
	return s
}

func checkBirdCollision(s State) bool {
	if !s.IsInAir {
		return false
	}
	for _, x := range s.Birds {
		if x >= 0 && x <= DinoHitboxEnd {
			return true
		}
	}
	return false
}

func hasConflict(positions []float64, x float64, safeZone float64) bool {
	for _, pos := range positions {
		if pos >= x-safeZone && pos <= x+safeZone {
			return true
		}
	}
	return false
}
