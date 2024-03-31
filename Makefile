ALL_NAMES := $(foreach item,$(wildcard */main.go),$(firstword $(subst /, ,$(item))))
BUILD_ARCHS := $(foreach arch,$(filter $(shell go env GOOS)%,$(shell go tool dist list)),$(lastword $(subst /, ,$(arch))))

.PHONY: show
show:
	@echo $(ALL_NAMES)
	@echo $(BUILD_ARCHS)

.PHONY: clean
clean:
	@echo $(@)
	@rm -rf build/

define BuildTask
.PHONY: $(1)-$(2)
$(1)-$(2):
	@echo $(1)-$(2)
	@go build -o build/$(2)/$(1)-zip $(1)/main.go

endef

$(foreach item,$(ALL_NAMES),$(foreach arch,$(BUILD_ARCHS),$(eval $(call BuildTask,$(item),$(arch)))))

.PHONY: all
all: $(foreach item,$(ALL_NAMES),$(foreach arch,$(BUILD_ARCHS),$(item)-$(arch)))
