The Linode plugin allows Packer to communicate with [Linode](https://www.linode.com/).

### Installation

To install this plugin, copy and paste this code into your Packer configuration, then run [`packer init`](https://www.packer.io/docs/commands/init).

```hcl
packer {
  required_plugins {
    linode = {
      version = ">= 1.0.1"
      source  = "github.com/linode/linode"
    }
  }
}
```


Alternatively, you can use `packer plugins install` to manage installation of this plugin.

```sh
$ packer plugins install github.com/linode/linode
```


### Components

#### Builders

- [linode](/packer/integrations/linode/linode/latest/components/builder/linode) - The Linode Builder creates [Linode Images](https://www.linode.com/docs/guides/linode-images/) for use on [Linode](https://www.linode.com/).

