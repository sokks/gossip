TASK_DIR=`go env GOPATH`/src/github.com/sokks/gossip/task2
LOG_DIR=$(TASK_DIR)/log
DATA_DIR=$(TASK_DIR)/data
task2:
	go clean github.com/sokks/gossip/task2
	go install github.com/sokks/gossip/task2
	mkdir -p $(LOG_DIR)
	mkdir -p $(DATA_DIR)
	`go env GOPATH`/bin/task2 $(LOG_DIR) $(DATA_DIR)
draw:
	`which python3` $(TASK_DIR)/draw_plot.py $(DATA_DIR)