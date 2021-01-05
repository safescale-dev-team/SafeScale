include ../common.mk

.PHONY:	cli/safescale cli/safescaled cli/gateway clean vet

all:	cli/safescale cli/safescaled cli/gateway

vet:
	@(cd safescale && $(MAKE) $@)
	@(cd safescaled && $(MAKE) $@)
	@(cd gateway && $(MAKE) $@)

cli/safescale:
	@printf "%b" "$(MAKE_TRACE)$(OK_COLOR)$(INFO_STRING) Building binary safescale, $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";
	@(cd safescale && $(MAKE))

cli/safescaled:
	@printf "%b" "$(MAKE_TRACE)$(OK_COLOR)$(INFO_STRING) Building binary safescaled, $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";
	@(cd safescaled && $(MAKE))

cli/gateway:
	@printf "%b" "$(MAKE_TRACE)$(OK_COLOR)$(INFO_STRING) Building binary gateway, $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";
	@(cd gateway && $(MAKE))

generate:
	@(cd safescale && $(MAKE) $(@))
	@(cd safescaled && $(MAKE) $(@))
	@(cd gateway && $(MAKE) $(@))

gensrc:
	@(cd safescale && $(MAKE) $(@))
	@(cd safescaled && $(MAKE) $(@))
	@(cd gateway && $(MAKE) $(@))

clean:
	@(cd safescale && $(MAKE) $(@))
	@(cd safescaled && $(MAKE) $(@))
	@(cd gateway && $(MAKE) $(@))
	