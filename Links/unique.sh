# Cleans files
# rm -f merged.md
# Merge all files
# cat *.md > temp.md
# Removes all spaces at EOL
fileName=$1
# sed 's/[[:blank:]]*$//' ${fileName} > ${fileName}_temp
# # Removes all duplicates and maintains order; adds two spaces at end
# awk '!seen[$0]++ {print $0"  "}' ${fileName}_temp > ${fileName}
# # Cleans files
# # rm -f temp.md temp_removed_space.md
# rm ${fileName}_temp

awk 'NF && !seen[$0]++ { print } ' ${fileName} > ${fileName}_temp
sed 's/^#/\n&/' ${fileName}_temp > ${fileName}
rm ${fileName}_temp