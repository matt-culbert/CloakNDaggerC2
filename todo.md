# Todo

## Server

### 12/23/2024

#### Issue: Implant management in the server
The server development work should shift focus to what it's like to manage implants. Things like when an implant checks in, 
when results are ready, when an operator pivots from one endpoint to the next. This should flow seamlessly.

##### Requirements
Going to use Dash Apps which renders an HTML webpage that can:
1) Be interacted with
2) Has multiple elements
3) Displays nodes, and selecting a node opens an options box
   4) This options box lets you run commands, see past output, see implant history

### 2/13/2025

#### Issue: Commands for the implant to run
The implant needs a set of commands configured and the ability to return the data. The client and server need the same 
set of commands.

##### Requirements
1) Configure the implant to read a directory or file and write to a file
2) The implant sends back this information in the same parsable manner currently used
3) Configure the server and client to have these 3 commands and return an error if a different one is selected
4) Ensure that the server properly reads and renders the command output

## Implant

### 2/13/2025

#### Issue: DNS and HTTP modes both need to be configured to run indefinity
Both the DNS and HTTP modes for the implant need to run indefinitely. Right now, HTTP waits for a command then processes 
and exits and DNS exits if no command is waiting, otherwise processes then exits. This needs to change.

##### Requirements
1) The implant needs a loop added to both methods to go infinitely
   2) Add exit conditions for if a debugger or inspector tool is attached
2) The loop should just start at the top of the implant, after a request is processed then sleep and repeat
   2) Ensure if no command is waiting that the implant goes back to the top of the loop cycle
