CC=g++

ifeq ($(DEBUG),yes)
	CXXFLAGS=-Wall -g
	LDFLAGS=-Wall -g
else
	CXXFLAGS=-Wall
	LDFLAGS=-Wall
endif

INCPATH=inc
SRCPATH=src
OBJPATH=obj
LIBPATH=lib
BINPATH=bin

OBJS=$(OBJPATH)/sub.o $(OBJPATH)/sub_wrap.o
OUT=$(LIBPATH)/sub.so

INCLUDES=-I ./$(INCPATH)

#Set this to your go installation directory
EXE=/usr/local/go/bin
export PATH := bin:$(PATH)

default: $(OUT)

$(OUT): $(OBJS)
	$(CC) $(LDFLAGS) -shared -o $@ $^

obj/sub_wrap.o: sub_wrap.cxx inc/sub.h
	$(CC) $(CXXFLAGS) $(INCLUDES) -fpic -c $< -o $@

obj/sub.o: src/sub.cpp inc/sub.h
	$(CC) $(CXXFLAGS) $(INCLUDES) -fpic -c $< -o $@


sub_wrap.cxx:
	swig -go -c++ -intgosize 64 -soname sub.so sub.swigcxx

.PHONY: clean cleanall

clean:
	rm -f $(OBJPATH)/.o

cleanall: clean
	rm -f $(OUT)
	rm -f .6
	rm -f .a
	rm -f .so
	rm -f .cxx
	rm -f .c

build:
	@echo "Building bindings..."
	$(EXE)go tool 6c -I /usr/local/go/pkg/linux_amd64/ -D _64BIT sub_gc.c
	$(EXE)go tool 6g sub.go
	$(EXE)go tool pack grc sub.a sub.6 sub_gc.6




install:
	@echo "Installing go package..."
	#Rename swig file so go install command does not try to reprocess it
	mv sub.swigcxx sub.notswig
	export GOPATH=$$HOME/go/; </span>
	$(EXE)go install
	mv sub.notswig sub.swigcxx

	@echo "Installing go shared lib..."
	sudo cp -f lib/libSub.so /usr/local/lib/
	sudo ldconfig