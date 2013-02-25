all:
	gccgo -Wall -Ofast -o turgenev turgenev.go state.go mechanics.go io.go search.go tables.go
clean:
	rm turgenev
