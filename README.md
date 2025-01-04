# giraartnetd

A simple, hacked together, tool I used for a single concert to control the venue lights from a GrandMA 3 lighting console, where the venue lighting is controlled by a GIRA X1 module.

The version tagged as v0.0.1 is the version used. I did some cleanup after, but those changes are untested, because I neither have access to a GIRA controller, nor a GrandMA 3 on a regular basis.

The updates are currently throttled to one update per second, I did this to prevent hammering the GIRA controller with every DMX update. If you want to change that, look at the loop at the end of `func main()` in `cmd/giraartnetd/main.go`.

## how to (hopefully) get set up

What you'll need:
* Basic knowledge how to compile a go application
* Basic JSON knowledge
* A machine with two network interfaces (one _might_ work, but I highly recommend two), and `curl` or another way to make an HTTP-POST request to the GIRA controller
* GIRA controller
    * Controller IP address (example: `192.168.0.100`), refered to as `$IP` below
    * access token (example: `CCFVnrRp6lgZFczBJDsL4YG27RXwwyW0`), refered to as `$TOKEN` below
        * you can create a token using username and password of the GIRA control module (refered to as `$USERNAME` and `$PASSWORD` below)
* ArtNet
    * A free IP address (example: `2.0.0.123`), refered to as `$ART_IP` below
    * The network with mask (example: `2.0.0.0/8`), refered to as `$ART_NET` below
    * A free network number/sub universe number combination (`0` and `4` in the example below)

The goal is to create a config file (`config.json`) for your setup:

```json
{
  "server": "https://$IP",
  "ignore_ssl": true,
  "token": "CCFVnrRp6lgZFczBJDsL4YG27RXwwyW0",
  "artnet": {
    "network": "2.0.0.0/8",
    "net": 0,
    "sub_uni": 4
  }
}
```

### Step 1: compile the binary

Clone the repository and build it with (I recommend to use `CGO_ENABLED=0` when compiling for a machine different machine)

```bash
CGO_ENABLED=0 go build ./cmd/giraartnetd
```

### Step 2: GIRA network

Connect one interfaces to the network with the GIRA control module.

If something does not work as expected, it might be helpful to look at the API documentation at https://partner.gira.de/data3/Gira_IoT_REST_API_v2_EN.pdf

#### Optional: register client, create a token

If you do now have a token, you can create it using the `/api/clients` endpoint, if everything works you'll get a JSON document that contains the token.

```
> curl -k -d '{ "client": "com.github.jpicht.giraartnetd" }' -u "$USERNAME:$PASSWORD" "https://$IP/api/clients"
{ "token": "..." }
```

#### Optional: Test connection

```bash
curl -k 'https://192.168.178.88/api/uiconfig?token=$TOKEN' > uiconfig.json
```

`uiconfig.json` should contain a JSON document with the UI configuration of the GIRA control module.

### Step 3: connect to the ArtNet network

* Connect the second interface to the lighting console network, as you would any other ArtNet device. Make sure the IPs and network mask are set up correctly for the machine and the lighting desk to communicate with each other.
* Set up a dedicated ArtNet universe (unique Net+SubNet numbers, refered to as `$NET` and `$SUB_UNI` later) in the lighting console

### Step 4: config.json

You should now have all information you need to create the `config.json` for you're environment.

I recommend to put it into a directory with the `giraartnetd` binary, this will be assumed below.

#### Optional: test ArtNet without GIRA

This assumes `giraartnetd` is in the current directory, also containing `config.json` and the `uiconfig.json` from the GIRA connection test above.

```bash
./giraartnetd --fake-uiconfig uiconfig.json
```

You should see a list of generated DMX channel assignments, example (with the `uiconfig.json` from the `data` directory):

```
time="2025-01-04 13:35:07.4980" level=info msg="skipping scene Rot"
CH 0.  1 [a003] Gruppe links/Brightness
CH 0.  2 [a00u] RGB Strahler links/Red
CH 0.  3 [a00v] RGB Strahler links/Green
CH 0.  4 [a00w] RGB Strahler links/Blue
```

You should see JSON output when you change the values of the channels on the lighting the desk.

### Step 5: start the bridge

When you're satisfied that you will not create an unintentional light show, for example by using the wrong network / sub universe number, start the program:

This assumes `giraartnetd` is in the current directory, also containing `config.json` created above.

```bash
./giraartnetd
```
