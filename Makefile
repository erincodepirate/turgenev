all:
	gccgo -Wall -O2 -o bin/turgenev src/*.go
clean:
	rm bin/turgenev
