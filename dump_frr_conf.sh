sudo vtysh -c 'sh run' | grep -v "^no " | grep -v "^frr " | grep -v "^Current configuration" | grep -v "^Building" | grep '\S' | grep -v "^hostname"
