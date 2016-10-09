EXE=./ircmap
LAYOUT?=circo

all: map.png

%.dot:
	$(EXE) > $@

%.png: %.dot
	dot -Tpng -K$(LAYOUT) $< -o $@

$(EXE):
	go build

clean:
	rm -f map.*
