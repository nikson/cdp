CC=gcc

FLAGS=-O3 -lm

all: histogram

histogram: histogram.go
	go build -o histogram histogram.go 

# histogram: histogram.c
#	 $(CC) histogram.c $(FLAGS)  -o histogram

clean:
	rm -f histogram
