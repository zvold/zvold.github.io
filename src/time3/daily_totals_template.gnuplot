set term png enhanced font ",14" size %WIDTH%,%HEIGHT% background rgb "#d5cdb6"

set key top right outside horizontal

set xtics rotate by 90 right
set ytics in mirror
set yrange [0:]
set label 1 "hours" at graph 0, screen 0.96

set xdata time
set timefmt "%Y-%m-%d"
set format x "%m-%d"
set xrange [%XMIN%:%XMAX%] # graph x boundaries (in seconds since epoch)

set style fill solid border -1

set boxwidth 43200 absolute # 50% of 1 day in seconds
set xtics 86400 # make one xtic every day
unset mxtics

set rmargin at screen 1
set lmargin 2.5

# manual stacking of boxes, work on top of work+rest
plot '-' using 1:(($2+$3)/3600) with boxes lc rgb "#489100" t column(3), \
	   '-' using 1:($2/3600) with boxes lc rgb "#0072d4" t column(2)

