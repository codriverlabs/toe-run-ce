FROM public.ecr.aws/amazonlinux/amazonlinux:2023

# Install perf and dependencies including Perl for flamegraph generation and jq
RUN dnf install -y perf procps-ng wget tar gzip perl jq && dnf clean all

# Install AWS aperf with architecture detection
RUN ARCH=$(uname -m) && \
    case $ARCH in \
        x86_64) APERF_ARCH="x86_64" ;; \
        aarch64) APERF_ARCH="aarch64" ;; \
        *) echo "Unsupported architecture: $ARCH" && exit 1 ;; \
    esac && \
    wget -O aperf.tar.gz "https://github.com/aws/aperf/releases/download/v0.1.18-alpha/aperf-v0.1.18-alpha-${APERF_ARCH}.tar.gz" && \
    tar -xzf aperf.tar.gz && \
    ls -la && \
    mv aperf-v0.1.18-alpha-${APERF_ARCH}/aperf /usr/local/bin/aperf && \
    chmod +x /usr/local/bin/aperf && \
    rm -rf aperf.tar.gz aperf-v0.1.18-alpha-${APERF_ARCH}

# Create entrypoint script
COPY aperf/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Copy common scripts
COPY aperf/send-profile.sh /usr/local/bin/send-profile.sh
RUN chmod +x /usr/local/bin/send-profile.sh

ENTRYPOINT ["/entrypoint.sh"]
