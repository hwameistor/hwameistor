from locust import FastHttpUser, task

class Apitest(FastHttpUser):
    @task
    def Apitest(self):
        self.client.get("/cluster/auth/info")
        self.client.get("/cluster/drbd")
        self.client.get("/cluster/events?page=1&pageSize=1")
        self.client.get("/cluster/operations?page=1&pageSize=1")
        self.client.get("/cluster/localdisknodes")
        self.client.get("/cluster/localdisks")
        self.client.get("/cluster/nodes?page=1&pageSize=1")
        self.client.get("/cluster/nodes/k8s-master")
        self.client.get("/cluster/nodes/k8s-master/disks?page=1&pageSize=1")
        self.client.get("/cluster/nodes/k8s-master/migrates?page=1&pageSize=1")
        self.client.get("/cluster/nodes/k8s-master/pools?page=1&pageSize=1")
        self.client.get("/cluster/pools?page=1&pageSize=1")
        self.client.get("/cluster/snapshots?page=1&pageSize=1")
        self.client.get("/cluster/volumegroups")
        self.client.get("/cluster/volumes?page=1&pageSize=1")