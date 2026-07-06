SWIG ?= /usr/bin/swig
RM=rm -f
CP=cp -f

SWIG_FLAG = -go -cgo -c++ -intgosize 64
ifeq ($(CONFIGURED_ARCH),arm64)
SWIG_FLAG += -DSWIGWORDSIZE64
endif
ifeq ($(CONFIGURED_ARCH),armhf)
SWIG_FLAG = -go -cgo -c++ -intgosize 32 -DSWIGWORDSIZE32
endif


.PHONY: all clean

all:
	$(CP) /usr/share/swss/swsscommon.i .
	# SWIG 4.4 (resolute) does not expand $function in %exception blocks for
	# the Go backend; use $action (standard, expands to the wrapped call).
	sed -i 's/\$$function/\$$action/g' swsscommon.i
	$(SWIG) $(SWIG_FLAG) -I/usr/include/swss/ swsscommon.i

clean:
	$(RM) *.go *.cxx *.i
