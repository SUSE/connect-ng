PAGES=SUSEConnect.8 SUSEConnect.5 zypper-migration.8 zypper-search-packages.8

all: $(PAGES)

%.8: %.8.md
	pandoc $^ -f markdown-smart-raw_html -t man -s -o $@

%.5: %.5.md
	pandoc $^ -f markdown-smart-raw_html -t man -s -o $@

clean:
	rm $(PAGES)
