# iterations per run (bench default is 100); example: just iter=20 run
iter := "100"
# node binary; example: just node=/usr/local/n/versions/node/22.23.3/bin/node run
node := "node"
# benchmarks: all, or comma list of floyd,lcs,quicksort; example: just bench=floyd run
bench := "all"
# languages: all, or comma list (js,bun,deno,scriptc,go,java,cs,cs-opt,c,rust,zig,swift); example: just lang=go,rust run
lang := "all"

# list recipes; example: just
default:
    @just --list

# build+run benchmarks sequentially, write <bench>.md; examples: just run | just bench=floyd run | just bench=floyd,lcs lang=go,rust iter=20 run
run:
    go run -C harness . -root {{justfile_directory()}} -bench {{bench}} -lang {{lang}} -iter {{iter}} -node {{node}}

# alias for run; example: just all | just iter=5 all
all: run

# remove build artifacts; example: just clean
clean:
    rm -rf build
