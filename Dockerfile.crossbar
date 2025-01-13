FROM crossbario/crossbar

USER root
COPY crossbar /mynode
RUN chown -R crossbar:crossbar /mynode

ENTRYPOINT ["crossbar", "start", "--cbdir", "/mynode/.crossbar"]
