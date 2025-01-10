<?php
namespace OPNsense\DHCPAdGuardSync;
class IndexController extends \OPNsense\Base\IndexController
{
    public function indexAction()
    {
        $this->view->pick('OPNsense/DHCPAdGuardSync/index');
    }
}