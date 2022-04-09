B1
What do you think will happen if you delete the pod running your service? Note down your prediction, then run kubectl delete pod <pod-name> (you can find the name with kubectl get pod) and wait a few seconds, then run kubectl get pod again. What happened and why do you think so?

It seems like the k8s scheduler would allocate another pod for my service. A new pod is created (or rescheduled?). As stated in the lab, "Deployments manage a set of Pods and ensure they keep running. If the Pods die or are removed for any reason the deployment will reschedule them."


B4
Welcome! You have chosen user ID 200485 (Walsh7839/ressiekonopelski@streich.biz)

Their recommended videos are:
 1. The bored mallard's childhood by Vida Doyle
 2. Crocodilehave: bypass by Ryder Aufderhar
 3. Sleepwrite: bypass by Floyd Turcotte
 4. The modern hedgehog's relaxation by Winona Ryan
 5. Sardinecan: navigate by Alexanne Morar


 B8
 Discussion: What does the load distribution look like with a client pool size of 4? What would you expect to happen if you used 1 client? How about 8? Note this down under section C3 in discussions.md.

 The graph traffic chart shows a smoother change in traffic, indicating the traffic is distributed acorss different connections. If only one client is used it is using the same pod thus would result in the same as before optimization. If 8  clients are used I would suspect the curve to be even smoother as less requests are taken by each pod now.