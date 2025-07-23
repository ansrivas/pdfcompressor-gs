# Maximum compression (will significantly reduce file size)
./pdftool compress large.pdf small.pdf 20

# Balanced compression 
./pdftool compress document.pdf compressed.pdf 60

# Light compression (maintains high quality)
./pdftool compress presentation.pdf optimized.pdf 85

./pdftool convert image.png output.pdf