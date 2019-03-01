HDRS = MatherDispatcher.h
C_SRC = mather.c MatherDispatcher.c
OUTPUT = mather.contract
LIBS = 
ABI_OUTPUT = MatherDispatcher.h MatherDispatcher.c
ABI_DEFS = Mather.abi

# tools to use
CC = i686-qtum-gcc
CXX = i686-qtum-g++
AR = i686-qtum-ar
AS = i686-qtum-as
LD = i686-qtum-ld
ABIGEN = ~/gopath/bin/SimpleABI

C_OBJS = $(subst .c,.o,$(C_SRC))

#default flags will just remove dead code and give warnings
CFLAGS += -Wall -ffunction-sections -fdata-sections
LDFLAGS += -Wl,--gc-section


default: build

build: $(OUTPUT)

$(OUTPUT): $(C_OBJS)
    $(CC) $(LDFLAGS) -o $(OUTPUT) $(C_OBJS) $(LIBS)

$(C_OBJS): $(HDRS) $(C_SRC) $(ABI_OUTPUT)
    $(CC) $(CFLAGS) -c $*.c -o $@

$(ABI_OUTPUT): $(ABI_DEFS)
    # hack, abi_defs should be capable of multiple files
    $(ABIGEN) --decode --abi $(ABI_DEFS)

clean:
    rm -f $(C_OBJS) $(OUTPUT)