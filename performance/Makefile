EXPERIMENT=1
N_OF_NODES=50
BASE_PORT=9080
MIN_DEGREE=5
MAX_DEGREE=7
TTL=10
LOG_FILEPATH=
install:
	go install github.com/sokks/gossip/performance
task2:
	go install github.com/sokks/gossip/task2
	`go env GOPATH`/bin/task2 $(LOGFILEPATH)
run:
	`go env GOPATH`/bin/performance $(N_OF_NODES) $(BASE_PORT) $(MIN_DEGREE) $(MAX_DEGREE) $(TTL)