DEPLOY ?= qtdeploy
FLAGS ?= 

build:
	$(DEPLOY) $(FLAGS) build desktop .

test:
	$(DEPLOY) $(FLAGS) test desktop .

test-fast:
	$(DEPLOY) $(FLAGS) --fast test desktop .

clean:
	find . -type f -name 'moc.*' -exec rm {} +
	find . -type f -name 'moc_*' -exec rm {} +
	find . -type f -name 'rcc.*' -exec rm {} +
	find . -type f -name 'rcc_*' -exec rm {} +
	rm -r darwin
	rm -r deploy
