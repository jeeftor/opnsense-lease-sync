<?php
namespace OPNsense\Dhcpsync;
class GeneralController extends \OPNsense\Base\IndexController
{
    public function indexAction()
    {
        // pick the template to serve to our users.
        $this->view->pick('OPNsense/Dhcpsync/general');
        $this->view->generalForm = $this->getForm("general");

        // Set custom page title to avoid duplication
        $this->view->title = "DHCP AdGuard Sync Settings";
    }
}
