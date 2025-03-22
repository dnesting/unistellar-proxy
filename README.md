# unistellar-proxy

This is a simple proxy server that forwards requests to a Unistellar telescope on another network.

## The Problem

The Unistellar eVscope is a great telescope, but it has a major limitation: it can only be controlled from its own WiFi network.
This means you have to be physically close to the telescope, and you can only control one telescope at a time.

## Usage

```sh
unistellar-proxy
```

The simple invocation assumes that the telescope is reachable at 192.168.100.1 (the IP address it uses on its private WiFi network), and that other clients can find the proxy on the network they are on.  This implies you have a system straddling (but not forwarding) traffic on both networks.

## Typical hardware setup

1. A device (e.g. Raspberry Pi) that can be simultaneously connected to the telescope's WiFi network and your home network.
2. This proxy server running on the device.

### Example setup (Arch Linux on Raspberry Pi)

#### Setup WiFi

Example `/etc/wpa_supplicant.conf`:

```
network={
  ssid="eVscope-12345-(Example)"
  key_mgmt=NONE
}
```

Commands:

```sh
systemctl enable dhcpcd@wlan0
systemctl start dhcpcd@wlan0
```

#### Build the proxy server

```sh
make
```

Install the resulting `unistellar-proxy` binary to `/usr/local/bin`.  You can run it by hand at this point to see if it works.

### Setup Systemd service

Create `/etc/systemd/system/unistellar-proxy.service` from [systemd/unistellar-proxy.service](systemd/unistellar-proxy.service).  Then:

```
systemctl daemon-reload
systemctl enable unistellar-proxy
systemctl start unistellar-proxy
```

## Alternative configurations

There's no reason you couldn't run this over the internet in a more complex configuration, for instance:

1. Local system, with wlan0 connected to the telescope (192.168.100.1) and en0 connected to your local network (10.0.0.2).  This node would run `unistellar-proxy` with no arguments and expose the telescope on 10.0.0.2 to other devices on the 10.0.0.0 network.
2. Remote system, with en0 (10.5.5.5) that can reach the local system (10.0.0.2).  You'd run `unistellar-proxy -proxy-to=10.0.0.2` to advertise the telescope remotely and forward requests to the local proxy.

None of this has been tested.
