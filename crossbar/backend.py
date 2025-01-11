from autobahn.twisted.wamp import ApplicationSession


class BackendSession(ApplicationSession):
    async def onJoin(self, details):
        print(f"Backend session joined: {details}")

        # SUBSCRIBE to a topic
        topic = "io.xconn.backend.topic1"

        def onhello(msg):
            print(f"event received on {topic}: {msg}")

        await self.subscribe(onhello, topic)
        print(f"subscribed to topic {topic}")

        # REGISTER a procedure for remote calling
        def add2(x, y):
            print(f"add2() called with {x} and {y}")
            return x + y

        await self.register(add2, "io.xconn.backend.add2")
        print("procedure add2() registered")

    def onLeave(self, details):
        print("Backend session leaved")
