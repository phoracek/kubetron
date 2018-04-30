package cniplugin

/*

actual functionality here

listen on server iface

mark node with label


get pod info from kubernetes
create iface on br-int, move to to netns
call dhclient ipam on it
done


on shutdown, remove label

*/

/*
TODO: either use sytem auth or run as a daemon
with endpoints in system (would allow us for graceful
removal)

TODO how do i get Pod ID from here?

containerID = k8s_....name_namespace

MUST use atomic operations to install it

install this as a daemon set

mount cni config and plugin dir
(configuration done in manifest, mount from wherever)

install ovn plugin and multus

inject default config within

*/

// TODO: first implement installation and client though
/*

//
//
//

type pluginServer struct {
}

// TODO: implement add and del, just minimal to try if it is triggered
// TODO: then read from pod
// TODO: then actually attach

//
//
//

const defaultBrName = "cni0"

func init() {
	// this ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
}

func cmdAdd(args *skel.CmdArgs) error {
	podName := ""
	podNamespace := ""

	networksSpec := nil

	// TODO: iterate spec and add ifaces

	err := exec.Command(
		"ovs-vsctl", "--",
		"add-port", integrationBridge, spec.PortName, "--",
		"set", "Interface", spec.PortName, "type=internal", fmt.Sprintf("external_ids:iface-id=%s", spec.PortName),
	).Run()

	//
	//
	//

	// TODO: no need to read that
	n, cniVersion, err := loadNetConf(args.StdinData)
	if err != nil {
		return err
	}

	br, brInterface, err := setupBridge(n)
	if err != nil {
		return err
	}

	// TODO: print args.Netns, maybe it is all we need
	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer netns.Close()

	// TODO: call dhcp ipam all the time
	// TODO: but disable default route and such stuff
	r, err := ipam.ExecAdd(n.IPAM.Type, args.StdinData)
	if err != nil {
		return err
	}

	// Convert whatever the IPAM result was into the current Result type
	result, err := current.NewResultFromResult(r)
	if err != nil {
		return err
	}

	if len(result.IPs) == 0 {
		return errors.New("IPAM plugin returned missing IP config")
	}

	// TODO: what is this? list all added ifaces
	result.Interfaces = []*current.Interface{brInterface, hostInterface, containerInterface}

	// TODO: also use result.IPs

	return types.PrintResult(result, cniVersion)
}

func cmdDel(args *skel.CmdArgs) error {
	n, _, err := loadNetConf(args.StdinData)
	if err != nil {
		return err
	}

	if err := ipam.ExecDel(n.IPAM.Type, args.StdinData); err != nil {
		return err
	}

	if args.Netns == "" {
		return nil
	}

	// There is a netns so try to clean up. Delete can be called multiple times
	// so don't return an error if the device is already removed.
	// If the device isn't there then don't try to clean up IP masq either.
	var ipnets []*net.IPNet
	err = ns.WithNetNSPath(args.Netns, func(_ ns.NetNS) error {
		var err error
		ipnets, err = ip.DelLinkByNameAddr(args.IfName)
		if err != nil && err == ip.ErrLinkNotFound {
			return nil
		}
		return err
	})

	if err != nil {
		return err
	}

	if n.IPMasq {
		chain := utils.FormatChainName(n.Name, args.ContainerID)
		comment := utils.FormatComment(n.Name, args.ContainerID)
		for _, ipn := range ipnets {
			if err := ip.TeardownIPMasq(ipn, chain, comment); err != nil {
				return err
			}
		}
	}

	return err
}

func main() {
	skel.PluginMain(cmdAdd, cmdDel, version.All)
}
*/
