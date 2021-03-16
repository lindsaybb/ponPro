# ponPro
Golang-based Profile Display &amp; Editing Tool for the Iskratel PON Products which extends the lindsaybb/gopon library.

This is a client-side library that depends on connection to the GPON or XGS-PON OLT from Iskratel to complete. This package is being made public to demonstrate the intent of this project, and to expand the use-case demonstrated in the lindsaybb/gopon/cmd demo. The goal of this package is to combine the restconf functionality with a Terminal User Interface (TUI, see github.com/rivo/tview and github.com/gdamore/tcell) to enable granular inpsection and modification of the details and operations available from the Restconf API. The video linked below demonstrates the essential terminal handling of this package while explaining the elements of the PON profiles as they are being modified.

https://youtu.be/VGNKlGTWLcM
