#
# This runs the games service in a local process exposing it on port 9001.
#
run-game-service: frodo
	go run games/cmd/main.go

#
# This runs the score service in a local process exposing it on port 9002.
#
run-score-service: frodo
	go run scores/cmd/main.go

#
# This runs the sample consumer program that utilizes both service clients to make
# requests to the game and score services. This will fail if you do not have 'run-game-service'
# and 'run-score-service' already running.
#
run:
	go run main.go

#
# Generates the frodo artifacts for all of the services.
#
frodo:
	frodo gateway games/game_service.go && \
	frodo client  games/game_service.go && \
	frodo gateway scores/score_service.go && \
	frodo client  scores/score_service.go
