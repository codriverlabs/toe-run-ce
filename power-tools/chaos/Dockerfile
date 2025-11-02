FROM public.ecr.aws/amazonlinux/amazonlinux:2023

# Install stress-ng, gdb and other utilities
RUN yum install -y stress-ng procps-ng gdb && yum clean all

# Create chaos scripts directory
RUN mkdir -p /chaos

# Copy chaos scripts
COPY chaos/entrypoint.sh /entrypoint.sh
COPY chaos/process-chaos.sh /chaos/process-chaos.sh
COPY chaos/cpu-chaos.sh /chaos/cpu-chaos.sh
COPY chaos/storage-chaos.sh /chaos/storage-chaos.sh
COPY chaos/network-chaos.sh /chaos/network-chaos.sh
COPY chaos/memory-chaos.sh /chaos/memory-chaos.sh
COPY chaos/send-profile.sh /usr/local/bin/send-profile.sh

# Make scripts executable
RUN chmod +x /entrypoint.sh /chaos/*.sh /usr/local/bin/send-profile.sh

ENTRYPOINT ["/entrypoint.sh"]
