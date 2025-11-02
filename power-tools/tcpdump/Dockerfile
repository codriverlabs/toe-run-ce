FROM public.ecr.aws/amazonlinux/amazonlinux:2023

# Install tcpdump and dependencies
RUN dnf install -y tcpdump jq && dnf clean all

# Create entrypoint script
COPY tcpdump/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Copy common scripts
COPY tcpdump/send-profile.sh /usr/local/bin/send-profile.sh
RUN chmod +x /usr/local/bin/send-profile.sh

ENTRYPOINT ["/entrypoint.sh"]
