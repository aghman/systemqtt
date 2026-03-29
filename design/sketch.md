# Sketch

## Overview

The goal of this service is to provide a way to publish systemd unit status to MQTT.
It will be a simple service that will listen for systemd unit status changes and publish them to MQTT.

## Design

### Architecture

The service will be a simple go service that will listen for systemd unit status changes and publish them to MQTT.

### Implementation

The service will be a simple go service that will listen for systemd unit status changes and publish them to MQTT.

- Uses spf13/cobra for command line parsing
- Uses viper for configuration
- Uses zap for logging
- Uses Makefles for building


### Configuration
- The service can be configured via a configuration file, that can be either in JSON or YAML format.  
- The configuration file location can be passed in as a command line parameter or via environment variables.
- All configuration options must be available as evironment variables, configuration file options, and command line parameters.
- All configuration options must be validated and error out if the configuration is invalid.
- All configuration options must have a default value.
- All configuration options must have a description.

### Configuration Options
- agent - format CLI output for AI agents, which equates to outputting everything in JSON format
- config.file - The location of the configuration file.
- config.format - The format of the configuration file.  Must be one of: json, yaml.
- mqtt.broker.url - The URL of the MQTT broker.
- mqtt.broker.port - The port of the MQTT broker.
- mqtt.broker.username - The username for the MQTT broker.
- mqtt.broker.password - The password for the MQTT broker.
- mqtt.client.id - The client ID for the MQTT broker.
- systemd.unit.filter - A comma separated list of systemd unit names to filter.  If not set, all units will be published.
- verbose - The verbosity of the output.  Must be one of: debug, info, warning, error, fatal.

### Commands
- config - Base command for handling configuration. Prints the current known configuration options and their values.
- config.template - Outputs a template configuration file with all configuration options and their default values.
- config.validate - Validates the configuration file, prints success or failure, and identifies any invalid options.
- doctor - Runs validation of configuration (from all known sources), prints success or failure, and identifies any invalid options.  Also attempts to validate connection to the MQTT broker.
- doctor.fix - Attempts to fix any invalid options in the configuration file.
- serve - Starts the service and listens for systemd unit status changes.
- version - Prints the version of the service.


## Runtime considerations
- The service will run as a systemd unit.
- The service should be run as a user that has access to the systemd unit status.
