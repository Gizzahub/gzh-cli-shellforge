# Count statements directly: rounded function summaries cannot enforce a boundary.
NR == 1 {
    if ($0 !~ /^mode: (set|count|atomic)$/) { invalid = 1; exit 1 }
    next
}
NF != 3 || $2 !~ /^[0-9]+$/ || $3 !~ /^[0-9]+$/ {
    invalid = 1
    exit 1
}
{
    total += $2
    if ($3 > 0) covered += $2
}
END {
    if (invalid || NR < 2 || total == 0) {
        print "Invalid or empty coverage profile" > "/dev/stderr"
        exit 1
    }
    printf "Statement coverage: %d/%d = %.6f%% (required >= 80%%)\n", covered, total, 100 * covered / total
    if (covered * 100 < total * 80) exit 1
}
