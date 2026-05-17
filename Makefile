.PHONY: test bundle clean

test:
	$(MAKE) -C hermes test

bundle:
	$(MAKE) -C hermes bundle

clean:
	$(MAKE) -C hermes clean
