<?php
namespace OPNsense\Dhcpsync;

use OPNsense\Base\BaseModel;

class Dhcpsync extends BaseModel
{
    /**
     * Check if the module is enabled
     * @return bool
     */
    public function isEnabled()
    {
        if (isset($this->general) && isset($this->general->Enabled)) {
            return (string)$this->general->Enabled === '1';
        }
        return false;
    }
}
