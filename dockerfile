# Using rockylinux:8-minimal as the base image for GLIBC_2.28
FROM rockylinux:8-minimal

# Set environment variables
ENV LANG=en_US.UTF-8
ENV LC_ALL=en_US.UTF-8

#Install necessary dependencies
RUN microdnf install go

# Set the working directory inside the container
WORKDIR /app

# Default command
CMD ["/bin/bash"]