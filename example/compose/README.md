# Compose Service Gateways Example

This shows you how you can run multiple service gateways in a single
process/server. There are 2 common use-cases for this functionality:

* When doing local development, it's often convenient to have a single
  process that runs all 15 (or however many) of your services rather
  than making devs manage up/downs for 15 separate processes.
* To reduce costs when deploying to production. Early on in your product
  it may make sense to deploy your app as a monolith while you get
  revenue. Then as you scale, you can start to give individual services
  their own processes. For example, maybe you start off running A, B, and C in the
  same process, but as you grow you give B its own process while A and C
  still share a process.
  
The idea is that you get the code clarity/maintainability of a service-based
app, but you have the flexibility to run/deploy it in a way that makes
sense for your scale/complexity/budget.

We're just going to run the two services from the `multiservice/` and
the calculator service from the `basic/` example
here rather than define more services. In total, we'll run 3 separate
services in a single gateway/process.

```shell
make run
```

If everything is working properly, you should see the output:

```
10 + 3 = 13
Game 1 = Super Mario Bros.
High Score 1 = 20332 (Dog Man)
High Score 2 = 999 (Wolverine)
High Score 3 = 899 (Dog Man)
```

## What This Example Shows

* How you can run multiple services/gateways via a single HTTP server
* If you build to your service interfaces, you can interchange service
  clients and servers at will.
