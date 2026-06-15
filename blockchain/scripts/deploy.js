const hre = require("hardhat");

async function main() {
  console.log("Deploying LimsHashRegistry to", hre.network.name, "...");

  const [deployer] = await hre.ethers.getSigners();
  console.log("Deployer address:", deployer.address);

  const LimsHashRegistry = await hre.ethers.getContractFactory("LimsHashRegistry");
  const registry = await LimsHashRegistry.deploy();
  await registry.waitForDeployment();

  const address = await registry.getAddress();

  console.log("--------------------------------------------------");
  console.log("LimsHashRegistry deployed to:", address);
  console.log("--------------------------------------------------");
  console.log("Set CONTRACT_ADDRESS=" + address + " in adapter/.env");
  console.log("--------------------------------------------------");
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
