all:
	gccgo -Wall -O2 -o turgenev *.go
clean:
	rm turgenev
install:
	cp --remove-destination turgenev /usr/local/bin/turgenev
