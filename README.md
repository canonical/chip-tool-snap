## How to build
- Install ubunutu 22.04 lxc container in your host dev environment.
- Install snapcraft 7.x in the container.
- Ensure that you will have build-essentials installed in this container.
- Clone this repo.
- To build the snap use following command at bash prompt inside your container
  ``` snapcraft --destructive-mode ```

## Notes
- At this moment this will build for x86_64 arch only. For arm64, work is in progress. 
- You should have at least 35GB of disk space available.
- Since matter source code is very big, downloading whole source can take considerable amount of time, which affects snap build time.
- To test this snap
   - Install the snap
     - ``` snap install <snap name> ```
     - ``` snap connect <snap name>:bluez ```
  - May be you can try simple pairing operation over ethernet by using following command line once the snap is installed,
    - ``` chip-tool pairing ethernet 1 20202021 3840 <ip of device node> 5543 ```
    
   The above command will start matter controller node for server device node you can build and run  linux placeholder device app from matter repo.

- Finally check out the matter repo for more details. Starting point  can be following      
  [matter-build-guide](https://github.com/project-chip/connectedhomeip/blob/master/docs/guides/BUILDING.md)
