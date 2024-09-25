# Source code for the VHS go-mql-build example.
#
# To run:
#
#     vhs < go-mql.tape

Output assets/go-mql-successful.gif

Set Theme "catppuccin-mocha"
Set WindowBar Colorful
Set FontSize 24
Set Width 1920
Set Height 1080
Set Margin 20
Set MarginFill "#674EFF"
Set BorderRadius 10
Set TypingSpeed 100ms
Set FontFamily "JetBrainsMono Nerd Font Mono"
Set Shell zsh

Hide
Type@10ms 'cd "/home/assadkhan/.wine/drive_c/Program Files (x86)/MetaTrader 4 EXNESS/MQL4"'
Enter
Type@10ms 'clear'
Enter
Show

Type "go-mql-build-local -c Experts/MACDSampleSuccess.mq4"

Sleep 500ms

Enter

Sleep 7s
