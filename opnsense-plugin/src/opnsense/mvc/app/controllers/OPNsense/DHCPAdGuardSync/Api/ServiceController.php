<?php
namespace OPNsense\DHCPAdGuardSync\Api;

use OPNsense\Base\ApiControllerBase;
use OPNsense\Core\Backend;

class ServiceController extends ApiControllerBase
{
    public function statusAction()
    {
        $backend = new Backend();
        $response = $backend->configdRun("dhcpadguardsync status");
        return array("response" => $response);
    }

    public function startAction()
    {
        $backend = new Backend();
        $response = $backend->configdRun("dhcpadguardsync start");
        return array("response" => $response);
    }

    public function stopAction()
    {
        $backend = new Backend();
        $response = $backend->configdRun("dhcpadguardsync stop");
        return array("response" => $response);
    }

    public function restartAction()
    {
        $backend = new Backend();
        $response = $backend->configdRun("dhcpadguardsync restart");
        return array("response" => $response);
    }

    public function logsAction()
    {
        $backend = new Backend();
        $response = $backend->configdRun("dhcpadguardsync logs");
        return array("response" => $response);
    }

    public function testAction()
    {
        $backend = new Backend();
        $response = $backend->configdRun("dhcpadguardsync test");
        return array("response" => $response);
    }
}
