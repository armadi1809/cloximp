# ====== Compiler and Flags ======
CC      = clang
CFLAGS  = -Wall -Wextra -Wpedantic -std=c17 -O2
LDFLAGS =

# ====== Files ======
SRC = main.c chunk.c memory.c debug.c       
OBJ = $(SRC:.c=.o)                 
TARGET = clox                 


all: $(TARGET)


$(TARGET): $(OBJ)
	$(CC) $(OBJ) $(LDFLAGS) -o $(TARGET)
	@echo "✅ Build complete: $(TARGET)"


%.o: %.c
	$(CC) $(CFLAGS) -c $< -o $@
	@echo "🔧 Compiled: $<"


clean:
	rm -f $(OBJ) $(TARGET)
	@echo "🧹 Cleaned build files."

rebuild: clean all

.PHONY: all clean rebuild