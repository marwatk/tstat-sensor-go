# Wifi Thermostat Sensor Simulator

A go library and cli tool that simulates a Venstar ACC-TSENWIFIPRO remote temperature sensor.

These sensors use Google ProtoBuffers to communicate with the thermostat. See [message.proto](message.proto) for the discovered protocol.

The physical sensors work great, but they chew through batteries.
This allows me to use my existing sensor network to feed the thermostat an average temperature.
