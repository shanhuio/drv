package bosinit

const bashrc = `
PS1='\[\033[01;32m\]\h:\[\033[01;34m\]\w\[\033[00m\]\$ '

alias ls='ls --color=auto'
alias cdp='cd -P'

# enable programmable completion features.
if ! shopt -oq posix; then
  if [[ -f /usr/share/bash-completion/bash_completion ]]; then
    source /usr/share/bash-completion/bash_completion
  fi
fi
`

// BashRC provides a the .bashrc file.
var BashRC = &WriteFile{
	Path:        "/home/rancher/.bashrc",
	Permissions: FilePerm(0644),
	Owner:       "rancher",
	Content:     bashrc,
}
