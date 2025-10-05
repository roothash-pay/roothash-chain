// SPDX-License-Identifier: GPL-3.0
pragma solidity 0.8.24;

contract WrappedTW {
    uint8 public constant decimals = 18;

    mapping(address => uint256) internal _balanceOf;
    mapping(address => mapping(address => uint256)) internal _allowance;

    event Approval(address indexed src, address indexed guy, uint256 wad);
    event Transfer(address indexed src, address indexed dst, uint256 wad);
    event Deposit(address indexed dst, uint256 wad);
    event Withdrawal(address indexed src, uint256 wad);

    receive() external payable {
        deposit();
    }

    fallback() external payable {
        deposit();
    }

    function name() external view virtual returns (string memory) {
        return "Wrapped TW";
    }

    function symbol() external view virtual returns (string memory) {
        return "WTW";
    }

    function allowance(address owner, address spender) public view virtual returns (uint256) {
        return _allowance[owner][spender];
    }

    function balanceOf(address src) public view returns (uint256) {
        return _balanceOf[src];
    }

    function deposit() public payable virtual {
        _balanceOf[msg.sender] += msg.value;
        emit Deposit(msg.sender, msg.value);
    }

    function withdraw(uint256 wad) public virtual {
        require(_balanceOf[msg.sender] >= wad);
        _balanceOf[msg.sender] -= wad;
        payable(msg.sender).transfer(wad);
        emit Withdrawal(msg.sender, wad);
    }

    function totalSupply() external view returns (uint256) {
        return address(this).balance;
    }

    function approve(address guy, uint256 wad) public virtual returns (bool) {
        _allowance[msg.sender][guy] = wad;
        emit Approval(msg.sender, guy, wad);
        return true;
    }

    function transfer(address dst, uint256 wad) external returns (bool) {
        return transferFrom(msg.sender, dst, wad);
    }

    function transferFrom(address src, address dst, uint256 wad) public returns (bool) {
        require(_balanceOf[src] >= wad);

        uint256 senderAllowance = allowance(src, msg.sender);
        if (src != msg.sender && senderAllowance != type(uint256).max) {
            require(senderAllowance >= wad);
            _allowance[src][msg.sender] -= wad;
        }
        _balanceOf[src] -= wad;
        _balanceOf[dst] += wad;

        emit Transfer(src, dst, wad);

        return true;
    }
}
