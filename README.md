Terraform Provider WinMultScript V1.1.0
==================

Windows AMIs on Amazon do not come standard with cloud-init. This has caused me a ton of headache tying to pass multiple scripts and/or provide a mechanism for others to use my module, while "adding" or passing their scripts into it. I decided enough was enough and it was time to write a provider.

WinMultScript allows you to pass up to 5 scripts into another module. WinMultScript will need to be setup on the module you are invoking and you will pass your filepath(s) (relative to YOUR module or the module itself if you're running WinMultScript inside you're own module) and variables into that module as parameters.

Still to be done

- Implement a size check

How it works
------------

* WinMultScript uses Terraform's [compact()](https://www.terraform.io/docs/configuration/functions/compact.html) method to clean up the list of filepaths first
* Before processing the data, we initialize a string starting with a ```<powershell>``` xml tag
* Loop the supplied files, read them in and copy into the string we initialized
* Once completed looping, end the string with ```</powershell>``` to signal EOF
* WinMultScript then renders the whole string and returns the rendered data via the ```.render``` property
* Your ```aws``` provider handles the rest when you assign the results to the ```user-data``` property

Requirements
------------

* [Terraform](https://www.terraform.io/downloads.html) 0.11.x
* [Go](https://golang.org/doc/install) 1.11 (to build the provider plugin)

Usage
---------------------

Building The Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/terraform-providers/terraform-provider-WinMultScript`

```powershell
mkdir -p $env:GOPATH/src/github.com/terraform-providers
cd $env:GOPATH/src/github.com/terraform-providers
git clone https://github.com/Azayzel/terraform-provider-winmultiscript.git
go get
go install
go build -o "terraform-provider-WinMultScript_v1.1.0_x4.exe"
```

Now, move the built executable into your ```./terraform/plugins/<your_build>/``` folder in your TF project.
You can now re-run ```terraform init``` to [Sideload](https://www.terraform.io/docs/configuration/providers.html#third-party-plugins) WinMultiScript provider

Writing your templates/ scripts
------------------------------

* File extensions should be .tpl
* Do NOT add ```<powershell></powershell>``` tags or any other tags. Simply write your script as you would outside of TF

Example Script

```powershell

# Passed params from TF
$string = "${tpl_snow}";

# Due to TF 11 being restrictive in types, specify array in a string with ',' delimiter
$list = "${tpl_rain}";

"$(Get-Date) : Testing secondary data: $($string)" | Add-Content "C:\yourlog.log";

foreach($item in $list.split(",")){
    "$(Get-Date) : Testing list items: $($item)" | Add-Content "C:\yourlog.log"

}

```

When calling a module that is using WinMultiScript

```javascript

resource "module" "MyNewWindowsInstance" {
    ....
    // Secondary Scripts/ Vars
    secondary_scripts = [
        "${file("${path.module}/pathto/Script.tpl")}",
        "${file("${path.module}/pathto/another/Script.tpl")}"
    ]
    secondary_vars = {
        "tpl_snow" = "cold powdery white stuff",
        "tpl_rain" = "wet, from sky, makes things grow"
    }

    ...
```

Terraform ```main.tf``` Config

```javascript

  ...
  locals {

     // Added to allow for wincloudinit to work with lists (dumb Terraform v11 stuff)
    baseScript = ["${ file("${path.module}/templates/windows_base.ps1.tpl") }"]
    rebootScript = ["${ file("${path.module}/templates/windows_base_reboot.ps1.tpl") }"]

    // When passing in .tpl's into the module, we need the file to be read in like below
    secondaryScripts = ["${ file("${path.module}/templates/Do-Stuff.tpl") }"]

    // Concat all scripts
    winMultiScripts = ["${ concat(local.baseScript, local.secondaryScripts, local.rebootScript)}"]
}
...


    data "winmultiscript" "joined" {

    // The array of script contents
    content_list = "${ local.winMultiScripts }"

    base_vars {
                tpl_instance_name       = "${ var.instance_name }"
                tpl_domain              = "${ var.domain_name }"
                tpl_domain_fqdn         = "${ var.domain_fqdn }"
                tpl_user                = "test"
                tpl_user_pw             = "test2"
                tpl_ou                  = "${ var.target_ou }"
                tpl_instance_admin_sg   = "${ var.instance_admin_sg }"
                tpl_instance_login_sg   = "${ var.instance_login_sg }"
                tpl_instance_role_sg    = "${ var.instance_role_sg }"
    }

    // If passing vars with secondary scripts
    // example:
    secondary_vars = "${ var.secondary_vars }"
  }
```

Then in your ```"aws_instance"``` resource

```javascript
 resource "aws_instance" "windows_2016_base" {
    ...
   user_data             = "${ data.WinMultScript.joined.rendered }"
    ...
```