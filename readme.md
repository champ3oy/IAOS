# IAOS: Incident Alerting and On-call Scheduling

## Overview

IAOS (Incident Alerting and On-call Scheduling) is a simple system designed to streamline incident management processes within small teams. It provides essential features such as authentication, on-call scheduling, incident alerting via Slack, and incident management, enabling teams to effectively respond to and resolve incidents promptly.

## Features

### Authentication

An account is created for the team and the admin can created sub-accounts for the team members(users)

### On-call Scheduling

Efficiently manage on-call rotations for teams, ensuring that responsible personnel are available to address incidents promptly, even outside of regular working hours.

### Incident Alerting (via Slack)

Integration with Slack enables real-time incident alerting to designated channels or individuals, ensuring immediate awareness and swift response to critical situations.

### Incident Management

Team members can acknowlegde incidents, resolve them and add follow ups.

## Setup

To get started with IAOS, follow these steps:

1. **Prerequisites**: Ensure you have Go installed on your machine.

2. **Clone the Repository**: Clone the IAOS repository from GitHub.

   ```
   git clone https://github.com/your-organization/iaos.git
   ```

3. **Configuration**: Navigate to the project directory and configure IAOS by providing the required settings. You'll need to set up environment variables or a configuration file, specifying details such as database connection strings, Slack Webhook URL, etc.

4. **Build and Run**: Build the IAOS application and start the server.

   ```
   go build
   ./iaos
   ```

5. **User Management**: Set up team account, add team members and assign appropriate roles to team members for accessing IAOS functionalities.

6. **On-call Schedule**: Define on-call rotations for your team members to ensure seamless coverage for incident response.

## How to Use

Once IAOS is set up and configured, you can start leveraging its features:

1. **Incident Reporting**: When an incident occurs, report it in IAOS, providing details such as severity, description, and affected systems.

2. **Alerting**: IAOS will automatically trigger alerts via Slack, notifying the designated responders about the incident.

3. **Incident Resolution**: Collaborate within IAOS to manage and resolve incidents efficiently. Update the incident status, track progress, and communicate with stakeholders.

4. **Post-Incident Analysis**: After resolving an incident, conduct post-mortems within IAOS to analyze the root cause, identify areas for improvement, and document lessons learned.

[by cirlormx](https://x.com/cirlormx)
