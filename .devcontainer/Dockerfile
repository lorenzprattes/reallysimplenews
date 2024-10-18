FROM node:latest


RUN apt-get update && apt-get install -y zsh git wget
RUN wget https://go.dev/dl/go1.23.2.linux-amd64.tar.gz && \
    rm -rf /usr/local/go && \
    tar -C /usr/local -xzf go1.23.2.linux-amd64.tar.gz && \
    rm go1.23.2.linux-amd64.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"
SHELL ["/bin/zsh", "-c"]
EXPOSE 3000
