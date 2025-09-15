<?php
namespace OPNsense\DHCPAdGuardSync;

use OPNsense\Core\Config;

class SettingsController extends \OPNsense\Base\IndexController
{
    public function indexAction()
    {
        // Include form definitions
        $this->view->generalForm = $this->getForm("general");
        $this->view->settingsForm = $this->getForm("dialogSettings");

        // Pick the template to serve
        $this->view->pick('OPNsense/DHCPAdGuardSync/settings');
    }
}
